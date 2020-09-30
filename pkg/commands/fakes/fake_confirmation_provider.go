package fakes

type FakeConfirmationProvider struct {
	// return values for confirm request
	confirm bool
	err     error
	// tracks if confirmation was requested
	requested bool
}

func NewFakeConfirmationProvider(confirm bool, err error) *FakeConfirmationProvider {
	return &FakeConfirmationProvider{
		confirm: confirm,
		err:     err,
	}
}

func (f *FakeConfirmationProvider) Confirm(_ string, _ ...string) (bool, error) {
	f.requested = true
	return f.confirm, f.err
}

func (f *FakeConfirmationProvider) WasRequested() bool {
	return f.requested
}
