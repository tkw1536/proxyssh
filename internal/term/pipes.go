package term

import (
	"io"
	"os"
)

// Pipes can be used by a Process to implement piped input / output
type Pipes struct {
	StdoutPipe, StderrPipe *os.File
	StdinPipe              *os.File

	descriptors []io.Closer // will be closed after a call to Close()
}

// osPipe calls os.Pipe() and adds both ends to descriptors.
func (p *Pipes) osPipe() (*os.File, *os.File, error) {
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	p.descriptors = append(p.descriptors, pw, pr)
	return pr, pw, nil
}

// Stdout returns a pipe to Stdout
func (p *Pipes) Stdout() (stdout io.ReadCloser, err error) {
	stdout, p.StdoutPipe, err = p.osPipe()
	return
}

// Stderr returns a pipe to Stderr
func (p *Pipes) Stderr() (stderr io.ReadCloser, err error) {
	stderr, p.StderrPipe, err = p.osPipe()
	return
}

// Stdin returns a pipe to Stdin
func (p *Pipes) Stdin() (stdin io.WriteCloser, err error) {
	p.StdinPipe, stdin, err = p.osPipe()
	return
}

// ClosePipes closes all pipes (if any)
func (p *Pipes) ClosePipes() {
	for _, d := range p.descriptors {
		d.Close()
	}
	p.descriptors = nil
}
