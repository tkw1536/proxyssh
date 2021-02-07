package feature

import (
	"reflect"
	"testing"
)

func TestParseNetworkAddress(t *testing.T) {
	tests := []struct {
		name    string
		want    NetworkAddress
		wantErr bool
	}{
		{
			name: "localhost:8080",
			want: NetworkAddress{
				Hostname: "localhost",
				Port:     8080,
			},
			wantErr: false,
		},
		{
			name: "127.0.0.1:8080",
			want: NetworkAddress{
				Hostname: "127.0.0.1",
				Port:     8080,
			},
			wantErr: false,
		},
		{
			name: "[::1]:80",
			want: NetworkAddress{
				Hostname: "::1",
				Port:     80,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNetworkAddress(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNetworkAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseNetworkAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
