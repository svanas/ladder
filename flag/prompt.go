package flag

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"
)

// reads a line of input from a terminal without local echo. this is commonly used for inputting passwords and other sensitive data.
func prompt(name string) ([]byte, error) {
	var (
		err error
		tty *os.File
	)
	fd := uintptr(syscall.Stdin)
	if tty, err = os.OpenFile("/dev/tty", os.O_RDWR, 0666); err == nil {
		fd = tty.Fd()
	} else {
		tty = os.Stderr
	}
	fmt.Fprintf(tty, "Please paste your %s: ", name)
	var buf []byte
	if buf, err = term.ReadPassword(int(fd)); err != nil {
		return nil, err
	}
	fmt.Fprintln(tty)
	return buf, nil
}
