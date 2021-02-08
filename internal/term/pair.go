package term

import (
	"os"

	"github.com/creack/pty"
)

// Pair represents a pair of pty and tty
type Pair struct {
	pty, tty *Terminal
}

// Use instructs this Terminal pair to make use of pty
func (p *Pair) Use(pty *os.File) {
	p.pty = GetTerminal(pty)
	p.tty = nil
}

// Open creates a new Pair of pty and tty
func (p *Pair) Open(rawMode bool) error {
	pty, tty, err := pty.Open()
	if err != nil {
		return err
	}

	p.pty = GetTerminal(pty)
	p.tty = GetTerminal(tty)

	if rawMode {
		p.Raw()
	}

	return nil
}

// Raw places the inputs of this terminal into raw mode.
func (p *Pair) Raw() {
	p.pty.SetRawInput()
	p.tty.SetRawOutput()
	p.tty.SetRawInput()
}

// ResizeEvent indiciates a size to resize to.
type ResizeEvent struct {
	Width, Height uint16
}

// Handle is HandleWith(events, nil).
func (p *Pair) Handle(events <-chan ResizeEvent) {
	p.HandleWith(events, nil)
}

// HandleWith instructs this Pair to resize according to events in this channel.
// Whenever a resize occurs additionally occurs effect, unless it is nil.
func (p *Pair) HandleWith(events <-chan ResizeEvent, effect func(event ResizeEvent)) {
	if events == nil {
		return
	}

	go func() {
		for ev := range events {
			p.pty.ResizeTo(ev.Width, ev.Height)
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

// ExternalT returns a reference to the external terminal.
func (p *Pair) ExternalT() *Terminal {
	return GetTerminal(p.External())
}

// InternalT returns a reference to the internal terminal.
func (p *Pair) InternalT() *Terminal {
	return GetTerminal(p.Internal())
}

// UnhangHack will send a new line to the internal terminal if it has not yet been closed.
// This is to prevent any internal ReadLine() from blocking forever!
func (p *Pair) UnhangHack() {
	defer func() { recover() }() // ignore any error
	p.External().WriteString("\n")
}

// Restore restores the input modes of this terminal.
func (p *Pair) Restore() {
	if p == nil {
		return
	}

	p.tty.RestoreTerminal()
	p.pty.RestoreTerminal()
}

// Close restores the input modes of each terminal and then calls close.
// When p is not initialized, does nothing.
func (p *Pair) Close() {
	if p == nil {
		return
	}

	p.Restore()

	p.tty.Close()
	p.pty.Close()
}
