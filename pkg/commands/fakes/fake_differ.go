package fakes

type FakeDiffer struct {
	DiffResult string
	arg0       interface{}
	arg1       interface{}
}

func (fd *FakeDiffer) Diff(dOld, dNew interface{}) (string, error) {
	fd.arg0 = dOld
	fd.arg1 = dNew
	return fd.DiffResult, nil
}

func (fd *FakeDiffer) Args() (interface{}, interface{}) {
	return fd.arg0, fd.arg1
}
