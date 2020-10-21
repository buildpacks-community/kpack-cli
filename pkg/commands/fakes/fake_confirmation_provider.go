package fakes

import "github.com/pkg/errors"

type FakeConfirmationProvider struct {
	// return values for confirm request
	confirm bool
	err     error
	// tracks confirmation message, empty if unrequested
	requestedMsg string
}

func NewFakeConfirmationProvider(confirm bool, err error) *FakeConfirmationProvider {
	return &FakeConfirmationProvider{
		confirm: confirm,
		err:     err,
	}
}

func (f *FakeConfirmationProvider) Confirm(msg string, _ ...string) (bool, error) {
	f.requestedMsg = msg
	return f.confirm, f.err
}

func (f *FakeConfirmationProvider) WasRequestedWithMsg(msg string) error {
	if f.requestedMsg == "" {
		return errors.New("confirmation was not requested")
	}
	if f.requestedMsg != msg {
		return errors.Errorf("wrong confirmation message. expected: %v, actual: %v", msg, f.requestedMsg)
	}
	return nil
}

func (f *FakeConfirmationProvider) WasRequested() bool {
	return f.requestedMsg != ""
}
