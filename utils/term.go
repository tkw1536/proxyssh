package utils

import (
	"os"

	"github.com/moby/term"
)

// Code is this file is roughly adapted from https://github.com/docker/cli/blob/master/cli/streams/stream.go
// and other files in the same directory.
//
// These are licensed under the Apache 2.0 License.
// This license requires to state changes made to the code and inclusion of the original NOTICE file.
//
// The code was modified to be independent of the docker cli utility classes where applicable.
//
// The original license and NOTICE can be found below:
//
// Docker
// Copyright 2012-2017 Docker, Inc.
//
// This product includes software developed at Docker, Inc. (https://www.docker.com).
//
// This product contains software (https://github.com/creack/pty) developed
// by Keith Rarick, licensed under the MIT License.
//
// The following is courtesy of our legal counsel:
//
// Use and transfer of Docker may be subject to certain restrictions by the
// United States and other governments.
// It is your responsibility to ensure that your use and/or transfer does not
// violate applicable laws.
//
// For more information, please see https://www.bis.doc.gov
//
// See also https://www.apache.org/dev/crypto.html and/or seek legal counsel.

// NewWritePipe returns a new pipe where the write end is a terminal
func NewWritePipe() (read *os.File, write *Terminal, err error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return r, GetTerminal(w), nil
}

// NewReadPipe returns a new pipe where the read end is a terminal
func NewReadPipe() (read *Terminal, write *os.File, err error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return GetTerminal(r), w, nil
}

// Terminal represents a local terminal
type Terminal struct {
	file *os.File

	fd         uintptr
	isTerminal bool

	state *term.State
}

// GetTerminal gets the terminal corresponding to in.
func GetTerminal(file *os.File) *Terminal {
	var t Terminal
	t.file = file
	t.fd, t.isTerminal = term.GetFdInfo(file)
	return &t
}

// File returns the corresponding os.File
func (t *Terminal) File() *os.File {
	if t == nil {
		return nil
	}
	return t.file
}

// Close closes the underlying file descriptor
func (t *Terminal) Close() error {
	if t == nil || t.file == nil {
		return nil
	}
	return t.file.Close()
}

// IsTerminal checks if t is indeed a terminal
func (t *Terminal) IsTerminal() bool {
	return t.isTerminal
}

// SetRawInput puts t into raw input mode unless it is not a terminal
func (t *Terminal) SetRawInput() (err error) {
	if t == nil || !t.isTerminal || t.state != nil {
		return nil
	}
	t.state, err = term.SetRawTerminal(t.fd)
	return
}

// SetRawOutput puts t into raw output mode unless it is not a terminal
func (t *Terminal) SetRawOutput() (err error) {
	if t == nil || !t.isTerminal || t.state != nil {
		return nil
	}
	t.state, err = term.SetRawTerminalOutput(t.fd)
	return
}

// RestoreTerminal restores the terminal to previous values
func (t *Terminal) RestoreTerminal() error {
	if t == nil || t.state == nil {
		return nil
	}

	defer func() { t.state = nil }() // wipe state
	return term.RestoreTerminal(t.fd, t.state)
}

// ResizeTo resizes this terminal to a specific size
func (t *Terminal) ResizeTo(width, height uint16) {
	if !t.isTerminal {
		return
	}
	term.SetWinsize(t.fd, &term.Winsize{
		Width:  width,
		Height: height,
	})
}
