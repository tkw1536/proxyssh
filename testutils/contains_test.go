package testutils

import "testing"

func TestSliceContainsString(t *testing.T) {
	type args struct {
		slice []string
		str   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty slice contains nothing", args{nil, "test"}, false},
		{"first element is contained", args{[]string{"a", "b", "c"}, "a"}, true},
		{"second element is contained", args{[]string{"a", "b", "c"}, "b"}, true},
		{"third element is contained", args{[]string{"a", "b", "c"}, "c"}, true},
		{"non-element is not contained", args{[]string{"a", "b", "c"}, "d"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SliceContainsString(tt.args.slice, tt.args.str); got != tt.want {
				t.Errorf("SliceContainsString() = %v, want %v", got, tt.want)
			}
		})
	}
}
