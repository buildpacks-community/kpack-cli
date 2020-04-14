package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

type CredentialFetcher struct {
}

func (c CredentialFetcher) FetchPassword(envVar, prompt string) (string, error) {
	password := os.Getenv(envVar)
	if password != "" {
		return password, nil
	}

	_, err := fmt.Fprint(os.Stdout, prompt)
	if err != nil {
		return "", err
	}

	pwBytes, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", err
	}

	_, _ = fmt.Fprintln(os.Stdout, "")

	return string(pwBytes), nil
}

func (c CredentialFetcher) FetchFile(envVar, filename string) (string, error) {
	password := os.Getenv(envVar)
	if password != "" {
		return password, nil
	}

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
