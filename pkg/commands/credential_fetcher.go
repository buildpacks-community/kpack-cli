// Copyright 2020-2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

type CredentialFetcher struct {
}

func (c CredentialFetcher) FetchPassword(envVar, prompt string) (string, error) {
	password, ok := os.LookupEnv(envVar)
	if ok {
		return password, nil
	}

	_, err := fmt.Fprint(os.Stdout, prompt)
	if err != nil {
		return "", err
	}

	pwBytes, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	_, _ = fmt.Fprintln(os.Stdout, "")

	return string(pwBytes), nil
}
