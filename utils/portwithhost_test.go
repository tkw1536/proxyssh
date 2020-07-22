package utils

import (
	"reflect"
	"testing"
)

func TestParsePortWithHost(t *testing.T) {
	tests := []struct {
		name    string
		want    PortWithHost
		wantErr bool
	}{
		{
			name: "localhost:8080",
			want: PortWithHost{
				Host: "localhost",
				Port: 8080,
			},
			wantErr: false,
		},
		{
			name: "127.0.0.1:8080",
			want: PortWithHost{
				Host: "127.0.0.1",
				Port: 8080,
			},
			wantErr: false,
		},
		{
			name: "[::1]:80",
			want: PortWithHost{
				Host: "::1",
				Port: 80,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePortWithHost(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePortWithHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePortWithHost() = %v, want %v", got, tt.want)
			}
		})
	}
}
