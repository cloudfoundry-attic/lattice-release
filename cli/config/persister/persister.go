package persister

type Persister interface {
	Load(interface{}) error
	Save(interface{}) error
}
