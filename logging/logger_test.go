// Package logging provides Logger.
package logging

import (
	"fmt"
	"net"
	"testing"
)

// testLoggerObj that writes the log message into message
type testLoggerObj struct {
	message string
}

var _ Logger = (*testLoggerObj)(nil)

func (obj *testLoggerObj) Print(v ...interface{}) {
	obj.message = fmt.Sprint(v...)
}
func (obj *testLoggerObj) Printf(format string, v ...interface{}) {
	obj.message = fmt.Sprintf(format, v...)
}

// testSessionOrContext implements LogSessionOrContext for testing purposes
type testSessionOrContext struct{}

var _ LogSessionOrContext = (*testSessionOrContext)(nil)

func (obj testSessionOrContext) User() string {
	return "user"
}

func (obj testSessionOrContext) RemoteAddr() net.Addr {
	return testSessionAddr{}
}

// testSessionAddr implements net.Addr
type testSessionAddr struct{}

func (testSessionAddr) Network() string { return "tcp" }
func (testSessionAddr) String() string  { return "0.0.0.0:0" }

func TestFmtSSHLog(t *testing.T) {
	type args struct {
		message string
		args    []interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"print simple message",
			args{
				message: "string is %s",
				args:    []interface{}{"string"},
			},
			"[user@0.0.0.0:0] string is string",
		},

		{
			"print complex message",
			args{
				message: "%d + %d is %s",
				args:    []interface{}{1, 2, "3"},
			},
			"[user@0.0.0.0:0] 1 + 2 is 3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &testLoggerObj{}
			FmtSSHLog(logger, testSessionOrContext{}, tt.args.message, tt.args.args...)

			got := logger.message
			if got != tt.want {
				t.Errorf("FmtSSHLog(): Got %q, want %q", got, tt.want)
			}
		})
	}
}
