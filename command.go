package proxyssh

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/gliderlabs/ssh"
	"github.com/kr/pty"
	"github.com/pkg/errors"
	"github.com/tkw1536/proxyssh/utils"
)

// HandleShellCommand creates an ssh.Handler that runs a shell command for every ssh.Session that connects
//
// shellCommand is a function that is called for every session and should return the shell command to run along with all arguments.
// When err is not nil, the first value of the command is passed to exec.LookPath.
//
// logger is called for every significant event for every connection.
func HandleShellCommand(logger utils.Logger, shellCommand func(session ssh.Session) (command []string, err error)) ssh.Handler {
	return func(session ssh.Session) {
		// logging
		utils.FmtSSHLog(logger, session, "session_start %s", session.User())
		defer utils.FmtSSHLog(logger, session, "session_end")

		// find the command to run
		command, err := shellCommand(session)
		if err != nil {
			abortsession(logger, session, errors.Wrap(err, "Failed to find command"))
			return
		}

		sshcmd, err := newSSHCommand(logger, session, command)
		if err != nil {
			abortsession(logger, session, errors.Wrap(err, "Failed to create ssh command"))
			return
		}

		utils.FmtSSHLog(logger, session, "session_valid %s", strings.Join(command, " "))
		sshcmd.Run()
	}
}

// abortsession aborts a session with a given error message
func abortsession(logger utils.Logger, s ssh.Session, err error) {
	errmsg := err.Error()
	utils.FmtSSHLog(logger, s, "session_command %s", errmsg)
	io.WriteString(s.Stderr(), errmsg+"\n")
	s.Exit(255)
}

// sshCommand represents an ssh session that proxies an ssh pty
type sshCommand struct {
	ssh.Session // the underlying ssh.session

	// the underlying logger
	Logger utils.Logger

	// the command and pty it's running on
	cmd    *exec.Cmd
	cmdPty *os.File

	// finish everything once
	finish sync.Once
}

// newSSHCommand returns a new ssh command
// s is the underlying ssh session, command is the command to be run
func newSSHCommand(logger utils.Logger, session ssh.Session, command []string) (c *sshCommand, err error) {

	// Look for 'exe' in the path, bail out if you can't find it
	var exe string
	exe, err = exec.LookPath(command[0])
	if err != nil {
		err = errors.Wrapf(err, "Can't find %s in path", command[0])
		return
	}

	// make the command
	cmd := exec.Command(exe, command[1:]...)

	// create the command session
	c = &sshCommand{
		Session: session,
		Logger:  logger,
		cmd:     cmd,
	}
	return
}

// Run runs and waits for this ssh session
// when an error occurs, calls finalize()
func (c *sshCommand) Run() error {
	err := c.start()

	// if we failed to start the command, we raise an error message
	if err != nil {
		err = errors.Wrap(err, "start() return error")
		c.finalize(255, err)
		return err
	}

	// if the user session disconnects, exit immediatly
	go func() {
		<-c.Context().Done()
		c.finalize(255, nil)
		return
	}()

	// wait for the command, then exit the ssh session appropriatly
	code, err := c.wait()
	c.finalize(code, err)
	return err
}

// start starts this session
// never calls finalize()
func (c *sshCommand) start() (err error) {

	// start either a regular or pty session
	if _, _, isPty := c.Pty(); !isPty {
		err = c.startRegular()
	} else {
		err = c.startPty()
	}
	return
}

// startRegular runs a non pty command
func (c *sshCommand) startRegular() error {
	// create a pipe for stdout
	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "cmd.StdoutPipe() returned error")
	}
	go func() {
		defer stdout.Close()
		io.Copy(c, stdout)
	}()

	// create a pipe for stderr
	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "cmd.StderrPipe() returned error")
	}
	go func() {
		defer stderr.Close()
		io.Copy(c.Stderr(), stderr)
	}()

	// create a pipe for stdin
	stdin, err := c.cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "cmd.StdinPipe() returned error")
	}
	go func() {
		defer stdin.Close()
		io.Copy(stdin, c)
	}()

	// and start the command
	return c.cmd.Start()
}

// startPty runs everything when a pty is requested
func (c *sshCommand) startPty() error {

	// create a new command and setup the term environment variable
	ptyReq, winCh, _ := c.Pty()
	c.cmd.Env = append(c.cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))

	// start a pty for the terminal
	f, err := pty.Start(c.cmd)
	if err != nil {
		err = errors.Wrap(err, "pty.Start() returned error")
		return err
	}
	c.fmtLog("pty_start_success")

	// listen for window size changes
	go func() {
		for win := range winCh {
			c.setWinsize(win.Width, win.Height)
		}
	}()
	go func() {
		io.Copy(f, c) // input
	}()
	io.Copy(c, f) // output

	return nil
}

func (c *sshCommand) wait() (code int, err error) {
	// wait for the command
	err = c.cmd.Wait()
	code = 255

	// if we have a failure and it's not an exit code
	// we need to return an error
	_, isExitError := err.(*exec.ExitError)
	if err != nil && !isExitError {
		err = errors.Wrap(err, "cmd.Wait() returned non-exit-error")
		return
	}

	// exit the session with the exit code
	code = c.cmd.ProcessState.ExitCode()
	c.fmtLog("command_return %d", code)
	return code, nil
}

// setWinsize sets the window size of the pty
func (c *sshCommand) setWinsize(w, h int) {
	c.fmtLog("term_resize %d %d", w, h)
	syscall.Syscall(syscall.SYS_IOCTL, c.cmdPty.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

// finalize finalizes this SSHCommand session.
// This function can be safely called multiple times, in different goroutines.
// If the session was already finalized, this function does nothing.
// Finalizing a session means setting a status code and, if err is not nil, print it to the stderr of the session and the log.
func (c *sshCommand) finalize(status int, err error) {
	c.finish.Do(func() {
		// write error message to console
		if err != nil {
			io.WriteString(c.Stderr(), err.Error()+"\n")
		}

		// mark that we are finalized, and return
		if err == nil {
			c.fmtLog("session_exit %d", status)
		} else {
			c.fmtLog("session_exit %d %s", status, err.Error())
		}
		c.Exit(status)

		// kill the process in the background
		go c.killProcess()
	})
}

// killProcess tries to kill the process.
// silences all errors
func (c *sshCommand) killProcess() {
	// no process => return
	if c.cmd.Process == nil {
		return
	}

	// if something goes wrong, ignore it
	defer func() {
		if r := recover(); r != nil {
			c.fmtLog("command_kill_failure")
		}
	}()

	// kill the process
	if c.cmd.Process.Kill() == nil {
		c.fmtLog("command_kill")
		c.cmd.Process = nil
	}
}

// fmtLog is like FmtSSHLog, but for this session
func (c *sshCommand) fmtLog(message string, args ...interface{}) {
	utils.FmtSSHLog(c.Logger, c, message, args...)
}
