package proxyssh

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"
	"github.com/tkw1536/proxyssh/utils"
)

// HandleShellCommand creates an ssh.Handler that runs a shell command for every ssh.Session that connects.
//
// shellCommand is a function.
// It is called for every ssh session and should return the shell command along with any arguments to execute for the provided session.
// The returned array must be at least of length 1.
// The first argument will be passed to exec.LookPath.
// When shell command returns a non-nil error, no command will be executed and the session will be aborted.
//
// logger is called for every significant event that occurs.
//
// See also the CommandSession struct and NewCommandSession func.
func HandleShellCommand(logger utils.Logger, shellCommand func(session ssh.Session) (command []string, err error)) ssh.Handler {
	return func(session ssh.Session) {
		// logging
		utils.FmtSSHLog(logger, session, "session_start %s", session.User())
		defer utils.FmtSSHLog(logger, session, "session_end")

		command, err := shellCommand(session)
		if err != nil {
			abortsession(logger, session, errors.Wrap(err, "Failed to find command"))
			return
		}

		// TODO: Have (command, args) returned
		sshcmd, err := NewCommandSession(logger, session, command[0], command[1:])
		if err != nil {
			abortsession(logger, session, errors.Wrap(err, "Failed to create ssh command"))
			return
		}

		utils.FmtSSHLog(logger, session, "session_valid %s", strings.Join(command, " "))
		sshcmd.Run()
	}
}

// abortsession exits an SSH session with code 255.
// It also prints the error message to the user on STDERR.
func abortsession(logger utils.Logger, s ssh.Session, err error) {
	errmsg := err.Error()
	utils.FmtSSHLog(logger, s, "session_command %s", errmsg)
	io.WriteString(s.Stderr(), errmsg+"\n")
	s.Exit(255)
}

// CommandSession represents an ongoing ssh.Session executing a shell command.
type CommandSession struct {
	ssh.Session // the underlying ssh.session

	Logger utils.Logger

	// the command that the session runs on
	cmd *exec.Cmd

	// for finalization
	started  utils.OneTime
	finished utils.OneTime
}

// NewCommandSession creates a new command session to execute a shell command.
// It prepares all resources, but does not actually start the session.
//
// The command and arguments describe the process to be running in this session.
// command will be passed to exec.LookPath.
func NewCommandSession(logger utils.Logger, session ssh.Session, command string, args []string) (*CommandSession, error) {

	// exec.Command internally does use LookPath(), but doesn't return an error
	// Instead we explicitly call LookPath() to intercept the error

	exe, err := exec.LookPath(command)
	if err != nil {
		err = errors.Wrapf(err, "Can't find %s in path", command)
		return nil, err
	}

	return &CommandSession{
		cmd: exec.Command(exe, args...),

		Session: session,
		Logger:  logger,
	}, nil
}

var errAlreadyStarted = errors.New("CommandSession.Run(): Already started. ")

// Run runs this session and waits for it to complete.
// After a call to this function ssh.Session will have closed.
//
// If this session was already started, immediatly returns an error.
// If something goes wrong when starting the session also returns an error.
func (c *CommandSession) Run() error {
	if !c.started.Lock() {
		return errAlreadyStarted
	}

	if err := c.start(); err != nil {
		err = errors.Wrap(err, "Failed to start session")
		c.finalize(255, err)
		return err
	}

	// if the user session disconnects, exit immediatly
	go func() {
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
func (c *CommandSession) start() (err error) {

	// start either a regular or pty session
	if _, _, isPty := c.Pty(); !isPty {
		err = c.startRegular()
	} else {
		err = c.startPty()
	}
	return
}

// startRegular starts running a command that did not request a pty
func (c *CommandSession) startRegular() error {
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

// startPty starts a session that requested a pty
func (c *CommandSession) startPty() error {
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
			c.fmtLog("term_resize %d %d", win.Height, win.Width)
			pty.Setsize(f, &pty.Winsize{
				Rows: uint16(win.Height),
				Cols: uint16(win.Width),
			})
		}
	}()

	go io.Copy(f, c) // input
	io.Copy(c, f)    // output

	return nil
}

// wait waits for this session to finish
func (c *CommandSession) wait() (code int, err error) {
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

// finalize finalizes this SSHCommand session.
// This function can be safely called multiple times, in different goroutines.
// If the session was already finalized, this function does nothing.
// Finalizing a session means setting a status code and, if err is not nil, print it to the stderr of the session and the log.
func (c *CommandSession) finalize(status int, err error) {
	if !c.finished.Lock() {
		return
	}

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
}

// killProcess tries to kill the process.
// silences all errors
func (c *CommandSession) killProcess() {
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
func (c *CommandSession) fmtLog(message string, args ...interface{}) {
	utils.FmtSSHLog(c.Logger, c, message, args...)
}
