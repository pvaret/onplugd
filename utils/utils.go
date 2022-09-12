package utils

import (
	"encoding/hex"
	"os"
	"os/user"
	"strings"

	"golang.org/x/sys/unix"
)

// Hex2uint16 takes a hexadecimal number encoded at a string and returns its
// value as a uint16, or 0 if the string is not a valid hexadecimal number. If
// the hexadecimal number is larger than FFFF, only the leftmost digits are used
// to compute the return value.
func Hex2uint16(s string) uint16 {
	s = strings.TrimSpace(s)
	if len(s)%2 == 1 {
		s = "0" + s
	}

	h, err := hex.DecodeString(s)
	if err != nil {
		return 0
	}
	if len(h) == 0 {
		return 0
	}
	if len(h) == 1 {
		return uint16(h[0])
	}
	return uint16(h[0])<<8 + uint16(h[1])
}

func getHomeDir() string {
	home := os.Getenv("HOME")

	if home != "" {
		return home
	}

	u, err := user.Current()
	if err != nil {
		// Welp.
		return "path"
	}
	return u.HomeDir
}

// Expand expands the "~" in a path.
func Expand(path string) string {
	home := strings.TrimSuffix(getHomeDir(), "/")

	if path == "~" || strings.HasPrefix(path, "~/") {
		return home + strings.TrimPrefix(path, "~")
	}

	return path
}

// IsATerminal determines if the given file is a TTY.
func IsATerminal(f *os.File) bool {
	_, err := unix.IoctlGetTermios(int(f.Fd()), unix.TCGETS)
	return err == nil
}
