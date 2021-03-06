package proxyssh

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"
	"github.com/tkw1536/proxyssh/internal/asyncio"
	"github.com/tkw1536/proxyssh/internal/lock"
	"github.com/tkw1536/proxyssh/internal/term"
	"github.com/tkw1536/proxyssh/logging"
)

// Session represents an ongoing ssh.Session executing a Process
type Session struct {
	ssh.Session                // the underlying ssh.session
	Logger      logging.Logger // for logging

	detector logging.MemoryLeakDetector // for keeping track of memory leak

	Process Process // the process that this session should execute

	// for finalization
	started  lock.OneTime
	finished lock.OneTime
}

// Process represents a runnable object with input and output streams.
type Process interface {
	fmt.Stringer

	// all methods in this interface are called at most once.

	// Init initializes this process
	Init(ctx context.Context, detector logging.MemoryLeakDetector, isPty bool) error
	Start(detector logging.MemoryLeakDetector, Term string, resizeChan <-chan WindowSize, isPty bool) (*os.File, error)

	// Input / Output Streams
	Stdout() (io.ReadCloser, error)
	Stderr() (io.ReadCloser, error)
	Stdin() (io.WriteCloser, error)

	// Wait waits for the process and returns the exit code
	Wait(detector logging.MemoryLeakDetector) (int, error)

	// Cleanup is called to cleanup this process, usually to kill it.
	Cleanup() (killed bool)
}

// WindowSize represents the size of the window
type WindowSize = term.Size

var errAlreadyStarted = errors.New("Session.Run(): Already started. ")

// Run runs this session and waits for it to complete.
// After a call to this function ssh.Session will have closed.
//
// If this session was already started, immediatly returns an error.
// If something goes wrong when starting the session also returns an error.
func (c *Session) Run() error {
	if !c.started.Lock() {
		return errAlreadyStarted
	}

	c.detector = logging.NewLeakDetector()
	if logging.MemoryLeakEnabled {
		c.fmtLog("memory_leak_detector_enabled")
	}

	if err := c.start(); err != nil {
		err = errors.Wrap(err, "Failed to start process")
		c.finalize(255, err)
		return err
	}

	// if the user session disconnects, exit immediatly
	c.detector.Add("session: context cancel")
	go func() {
		defer c.detector.Done("session: context cancel")

		<-c.Context().Done()
		c.finalize(255, nil)
		return
	}()

	// else wait for the session to finish
	code, err := c.wait()
	c.finalize(code, err)
	return err
}

// start starts this session
func (c *Session) start() error {

	// initialize the process
	_, _, isPty := c.Pty()
	if err := c.Process.Init(c.Session.Context(), c.detector, isPty); err != nil {
		return errors.Wrap(err, "Failed to initialize process")
	}

	// start either a regular or pty session
	if isPty {
		return c.startPty()
	}

	return c.startRegular()
}

// startRegular starts running a command that did not request a pty
func (c *Session) startRegular() error {
	// create a pipe for stdout
	stdout, err := c.Process.Stdout()
	if err != nil {
		return errors.Wrap(err, "Failed to connect to STDOUT")
	}

	c.detector.Add("session: stdout")
	go func() {
		defer c.detector.Done("session: stdout")
		defer stdout.Close()
		asyncio.CopyLeak(c.Context(), c, stdout)
	}()

	// create a pipe for stderr
	stderr, err := c.Process.Stderr()
	if err != nil {
		return errors.Wrap(err, "Failed to connect to STDERR")
	}
	c.detector.Add("session: stderr")
	go func() {
		defer c.detector.Done("session: stderr")
		defer stderr.Close()
		asyncio.CopyLeak(c.Context(), c.Stderr(), stderr)
	}()

	// create a pipe for stdin
	stdin, err := c.Process.Stdin()
	if err != nil {
		return errors.Wrap(err, "Failed to connect to STDIN")
	}

	c.detector.Add("session: stdin")
	go func() {
		defer c.detector.Done("session: stdin")
		defer stdin.Close()
		asyncio.CopyLeak(c.Context(), stdin, c)
	}()

	// and start the command
	_, err = c.Process.Start(c.detector, "", nil, false)
	return err
}

// startPty starts a session that requested a pty
func (c *Session) startPty() error {
	// create a new command and setup the term environment variable
	ptyReq, winCh, _ := c.Pty()

	// create a channel for resizing the window
	resizeChan := make(chan WindowSize)
	c.detector.Add("session: winCh")
	go func() {
		defer c.detector.Done("session: winCh")
		for win := range winCh {
			// c.fmtLog("term_resize %d %d", win.Height, win.Width)
			resizeChan <- WindowSize{
				Height: uint16(win.Height),
				Width:  uint16(win.Width),
			}
		}
		close(resizeChan)
	}()

	f, err := c.Process.Start(c.detector, ptyReq.Term, resizeChan, true)
	if err != nil {
		return err
	}
	c.fmtLog("pty_start_success")

	c.detector.Add("session: input")
	go func() {
		defer c.detector.Done("session: input")
		asyncio.CopyLeak(c.Context(), f, c)
	}()

	c.detector.Add("session: output")
	go func() {
		defer c.detector.Done("session: output")
		asyncio.CopyLeak(c.Context(), c, f)
	}()

	return nil
}

// wait waits for this session to finish
func (c *Session) wait() (code int, err error) {
	code, err = c.Process.Wait(c.detector)
	if err == nil {
		c.fmtLog("command_return %d", code)
	} else {
		c.fmtLog("command_return_fail %s", err)
	}
	return
}

// finalize finalizes this SSHCommand session.
// This function can be safely called multiple times, in different goroutines.
// If the session was already finalized, this function does nothing.
// Finalizing a session means setting a status code and, if err is not nil, print it to the stderr of the session and the log.
func (c *Session) finalize(status int, err error) {
	if !c.finished.Lock() {
		return
	}

	// write error detection code
	if err != nil {
		io.WriteString(c.Stderr(), err.Error()+"\n")
	}

	// trigger the leak detector
	c.detector.Finish(c.Logger, c.Session)

	// mark that we are finalized, and return
	if err == nil {
		c.fmtLog("session_exit %d", status)
	} else {
		c.fmtLog("session_exit %d %s", status, err.Error())
	}
	c.Exit(status)

	// kill the process in the background
	go c.killProcess()
}

// killProcess attempts to kill the underlying process.
func (c *Session) killProcess() {
	res := c.Process.Cleanup()
	if res {
		c.fmtLog("command_kill")
	} else {
		c.fmtLog("command_kill_failure")
	}
}

// fmtLog is like FmtSSHLog, but for this session
func (c *Session) fmtLog(message string, args ...interface{}) {
	logging.FmtSSHLog(c.Logger, c, message, args...)
}
