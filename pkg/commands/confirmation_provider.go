package commands

import (
	"bufio"
	"fmt"
	"os"
)

type ConfirmationProvider interface {
	Confirm(message string, okayResponses ...string) (bool, error)
}

type confirmationProviderImpl struct {
	defaultOkayResponses []string
}

func NewConfirmationProvider() ConfirmationProvider {
	return confirmationProviderImpl{
		defaultOkayResponses: []string{"y", "Y", "yes", "YES"},
	}
}

func (s confirmationProviderImpl) Confirm(message string, okayResponses ...string) (bool, error) {
	_, err := fmt.Fprint(os.Stdout, message)
	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(os.Stdin)
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
