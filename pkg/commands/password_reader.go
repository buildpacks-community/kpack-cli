package commands

import (
	"fmt"
	"io"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

type PasswordReader struct {
}

func (p PasswordReader) Read(writer io.Writer, prompt, envVar string) (string, error) {
	password := os.Getenv(envVar)
	if password != "" {
		return password, nil
	}

	_, err := fmt.Fprint(writer, prompt)
	if err != nil {
		return "", err
	}

	pwBytes, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", err
	}
	return string(pwBytes), nil
}
