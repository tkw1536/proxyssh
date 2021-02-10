package term

import (
	"os"
	"sync"

	"github.com/creack/pty"
	"github.com/moby/term"
)

// Pair represents a pair of pty and tty
type Pair struct {
	pty, tty *terminal
}

// Use instructs this Terminal pair to make use of pty
func (p *Pair) Use(pty *os.File) {
	p.pty = newTerminal(pty)
	p.tty = nil
}

// Open creates a new Pair of pty and tty
func (p *Pair) Open(rawMode bool) error {
	pty, tty, err := pty.Open()
	if err != nil {
		return err
	}

	p.pty = newTerminal(pty)
	p.tty = newTerminal(tty)

	if rawMode {
		p.RawMode()
	}

	return nil
}

// Size indicates the size of a terminal
type Size struct {
	Width, Height uint16
}

// Handle is HandleWith(events, nil).
func (p *Pair) Handle(events <-chan Size) {
	p.HandleWith(events, nil)
}

// HandleWith instructs this Pair to resize according to events in this channel.
// Whenever a resize occurs, the effect function is called except when it is nil.
func (p *Pair) HandleWith(events <-chan Size, effect func(event Size)) {
	if events == nil {
		return
	}

	go func() {
		for ev := range events {
			p.pty.Resize(ev)
			if effect == nil {
				continue
			}
			effect(ev)
		}
	}()
}

// External returns an os.File pointing to the master terminal.
// This is the terminal that should be used for communication.
func (p *Pair) External() *os.File {
	if p == nil {
		return nil
	}
	return p.pty.File()
}

// Internal returns an os.File pointing to the slave terminal.
// This is the terminal that internal programs should run on.
func (p *Pair) Internal() *os.File {
	if p == nil {
		return nil
	}
	return p.tty.File()
}

// UnhangHack will attempt to emulate a user exit on the underlying terminal.
// This is to unblock any potentially stuck "Read()" call to it.
func (p *Pair) UnhangHack() {
	defer func() { recover() }() // ignore any error

	// It turns out that in practice these calls may or may not work.
	// however they do stop the memory leaks in regular exit cases.

	// TODO: Figure out how to deal well with the hangup case
	// where the terminal just exits.

	pty := p.External()
	pty.Write([]byte{
		3,          // ctrl+c (end of text)
		4,          // ctrl+d (end of transmission)
		byte('\n'), // newline
	})
}

// RawMode places the inputs of this terminal into raw mode.
func (p *Pair) RawMode() {
	p.pty.RawMode()
	p.tty.RawMode()
}

// RestoreMode restores the input modes of this terminal.
func (p *Pair) RestoreMode() {
	if p == nil {
		return
	}

	p.tty.RestoreMode()
	p.pty.RestoreMode()
}

// Close restores the input modes of each terminal and then calls close.
// When p is not initialized, does nothing.
func (p *Pair) Close() {
	if p == nil {
		return
	}

	p.RestoreMode()

	p.tty.Close()
	p.pty.Close()
}

// terminal represents a local terminal
type terminal struct {
	file *os.File

	fd         uintptr
	isTerminal bool

	l     sync.RWMutex // l protects state
	state *term.State
}

// newTerminal returns a new terminal object based on *os.File
func newTerminal(file *os.File) *terminal {
	var t terminal
	t.file = file
	t.fd, t.isTerminal = term.GetFdInfo(file)
	return &t
}

// File returns the underlying os.File
func (t *terminal) File() *os.File {
	if t == nil {
		return nil
	}
	return t.file
}

// Close closes the underlying terminal mode
func (t *terminal) Close() error {
	if t == nil || t.file == nil {
		return nil
	}
	return t.file.Close()
}

// IsTerminal checks if t is indeed a terminal
func (t *terminal) IsTerminal() bool {
	return t.isTerminal
}

// RawMode puts t into raw (input) mode
func (t *terminal) RawMode() (err error) {
	if t == nil {
		return nil
	}

	t.l.Lock()
	defer t.l.Unlock()

	if t == nil || !t.isTerminal || t.state != nil {
		return nil
	}
	t.state, err = term.SetRawTerminal(t.fd)
	return
}

// RestoreMode restores the previous (input) mode
func (t *terminal) RestoreMode() error {
	if t == nil {
		return nil
	}

	t.l.Lock()
	defer t.l.Unlock()

	if t.state == nil {
		return nil
	}

	defer func() { t.state = nil }() // wipe state
	return term.RestoreTerminal(t.fd, t.state)
}

// Resize resizes this terminal to a specific size
func (t *terminal) Resize(size Size) {
	if !t.isTerminal {
		return
	}
	term.SetWinsize(t.fd, &term.Winsize{
		Width:  size.Width,
		Height: size.Height,
	})
}
