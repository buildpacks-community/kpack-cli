// Copyright 2020-Present VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
)

type defaultConfirmationProvider struct {
	reader               io.Reader
	writer               io.Writer
	defaultOkayResponses []string
}

func NewConfirmationProvider() defaultConfirmationProvider {
	return defaultConfirmationProvider{
		reader:               os.Stdin,
		writer:               os.Stdout,
		defaultOkayResponses: []string{"y", "Y", "yes", "YES"},
	}
}

func (s defaultConfirmationProvider) Confirm(message string, okayResponses ...string) (bool, error) {
	if message == "" {
		return false, errors.New("confirmation message cannot be empty")
	}

	_, err := fmt.Fprint(s.writer, message)
	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(s.reader)
	scanner.Scan()
	response := scanner.Text()

	if len(okayResponses) == 0 {
		okayResponses = append(okayResponses, s.defaultOkayResponses...)
	}

	for _, okayResponse := range okayResponses {
		if response == okayResponse {
			return true, nil
		}
	}

	return false, nil
}
