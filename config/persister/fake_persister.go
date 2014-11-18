package persister

type fakePersister struct {
	err error
}

func NewFakePersister() Persister {
	return &fakePersister{}
}

func NewFakePersisterWithError(err error) Persister {
	return &fakePersister{err}
}

func (f *fakePersister) Load(_ interface{}) error {
	return f.err
}

func (f *fakePersister) Save(_ interface{}) error {
	return f.err
}
