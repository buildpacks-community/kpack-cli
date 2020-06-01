package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
)

type ConfirmationProvider interface {
	Confirm(message string, okayResponses ...string) (bool, error)
}

type defaultConfirmationProvider struct {
	reader               io.Reader
	writer               io.Writer
	defaultOkayResponses []string
}

func NewConfirmationProvider() ConfirmationProvider {
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
