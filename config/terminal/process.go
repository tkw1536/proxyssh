package terminal

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/tkw1536/proxyssh"
	"github.com/tkw1536/proxyssh/internal/term"
	"github.com/tkw1536/proxyssh/logging"
	"golang.org/x/crypto/ssh/terminal"
)

// REPLProcess represents a process that is run using a the shell on the current machine
type REPLProcess struct {
	WelcomeMessage string
	Prompt         string
	Loop           func(ctx context.Context, w io.Writer, read string) (exit bool, code int)

	terminal *term.Pair

	workerContextCancel func() // called to cancel the worker context

	exitCode   int           // exit code the repl loop returned.
	loopWaiter chan struct{} // close()d when the main loop exists.
}

// DefaultREPL is the default function that a REPLConfig uses when the user does not provide one.
//
// It exits unless the user types the word 'exit'.
var DefaultREPL = func(ctx context.Context, w io.Writer, input string) (exit bool, code int) {
	return input == "exit", 0
}

// Init initializes this process.
func (repl *REPLProcess) Init(ctx context.Context, detector logging.MemoryLeakDetector, isPty bool) error {
	if repl.Loop == nil {
		repl.Loop = DefaultREPL
	}

	repl.loopWaiter = make(chan struct{})
	repl.workerContextCancel = func() {}

	return nil
}

// Wait waits for the process and returns the exit code.
func (repl *REPLProcess) Wait(detector logging.MemoryLeakDetector) (code int, err error) {
	detector.Add("terminal: Wait")
	defer detector.Done("terminal: Wait")

	<-repl.loopWaiter
	return repl.exitCode, nil
}

// Cleanup cleans up this process
func (repl *REPLProcess) Cleanup() (killed bool) {
	repl.workerContextCancel()

	// unhang the Close() method and exit!
	repl.terminal.UnhangHack()
	repl.terminal.Close()

	return true
}

// String turns REPLProcess into a string
func (repl *REPLProcess) String() string {
	return "REPLProcess"
}

// Stdout returns a pipe to Stdout
func (repl *REPLProcess) Stdout() (io.ReadCloser, error) {
	return nil, errNotATTY
}

// Stderr returns a pipe to Stderr
func (repl *REPLProcess) Stderr() (io.ReadCloser, error) {
	return nil, errNotATTY
}

// Stdin returns a pipe to Stdin
func (repl *REPLProcess) Stdin() (io.WriteCloser, error) {
	return nil, errNotATTY
}

var errNotATTY = errors.New("tty was not allocated")

// Start starts this process
func (repl *REPLProcess) Start(detector logging.MemoryLeakDetector, Term string, resizeChan <-chan proxyssh.WindowSize, isPty bool) (*os.File, error) {
	if !isPty {
		return nil, errNotATTY
	}
	return repl.startPty(detector, Term, resizeChan, isPty)
}

//
// PTY implementation
//

func (repl *REPLProcess) startPty(detector logging.MemoryLeakDetector, Term string, resizeChan <-chan proxyssh.WindowSize, isPty bool) (*os.File, error) {
	// create a new terminal
	repl.terminal = &term.Pair{}
	if err := repl.terminal.Open(true); err != nil {
		return nil, err
	}

	// handle the channel for resizing
	repl.terminal.Handle(resizeChan)

	// make a context for the terminal loop
	var ctx context.Context
	ctx, repl.workerContextCancel = context.WithCancel(context.Background())

	detector.Add("terminal: pty loop")
	go repl.runLoopPty(ctx, detector)

	// and return the pair of the terminal
	return repl.terminal.External(), nil
}

func (repl *REPLProcess) runLoopPty(ctx context.Context, detector logging.MemoryLeakDetector) {
	defer detector.Done("terminal: pty loop")

	// isCancelled checks if the context has been canceled
	isCancelled := func() bool {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	}

	// create a new terminal and write the welcome message to the terminal.
	term := terminal.NewTerminal(repl.terminal.Internal(), repl.Prompt)
	io.WriteString(term, repl.WelcomeMessage+"\n")

	// Keep reading input and running the REPL loop
	// until the loop is stopped external.

	for !isCancelled() {

		// Read a line from the input.
		// This block even if the input is Close()d.
		input, err := term.ReadLine()

		// if the context was canceled while the Read() call happened, don't do anything!
		if isCancelled() {
			break
		}

		// input ended => 0
		if err == io.EOF {
			repl.exitCode = 0
			break
		}

		// something went wrong while reading => 255
		if err != nil {
			repl.exitCode = 255
			break
		}

		// do the loop code
		exit, code := repl.Loop(ctx, term, input)
		if exit {
			repl.exitCode = code
			break
		}
	}

	// close the underlying terminals and exit
	repl.terminal.Close()
	close(repl.loopWaiter)
}
