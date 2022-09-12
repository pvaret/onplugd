package utils

import (
	"os"
	"testing"
)

const testHome = "/home/test"

func Test_Hex2uint16(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want uint16
	}{
		{
			name: "wrong",
			args: args{s: "X"},
			want: 0,
		},
		{
			name: "odd length",
			args: args{s: "A"},
			want: 10,
		},
		{
			name: "space",
			args: args{s: "  01  "},
			want: 1,
		},
		{
			name: "8 bits",
			args: args{s: "ff"},
			want: 255,
		},
		{
			name: "16 bits",
			args: args{s: "ff11"},
			want: 255<<8 + 17,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Hex2uint16(tt.args.s); got != tt.want {
				t.Errorf("hex2uint16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getHomeDir(t *testing.T) {
	want := testHome
	if got := getHomeDir(); got != want {
		t.Errorf("getHomeDir() = %v, want %v", got, want)
	}
}

func Test_Expand(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{path: ""},
			want: "",
		},
		{
			name: "tilde",
			args: args{path: "~"},
			want: "/home/test",
		},
		{
			name: "full",
			args: args{path: "~/test"},
			want: "/home/test/test",
		},
		{
			name: "no change",
			args: args{path: "/a/b/c"},
			want: "/a/b/c",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Expand(tt.args.path); got != tt.want {
				t.Errorf("expand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMain(m *testing.M) {
	os.Setenv("HOME", testHome)
	os.Exit(m.Run())
}
