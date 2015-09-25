package models

func (t *ModificationTag) Increment() {
	t.Index++
}

func (m *ModificationTag) SucceededBy(other *ModificationTag) bool {
	if m == nil || m.Epoch == "" || other.Epoch == "" {
		return true
	}

	return m.Epoch != other.Epoch || m.Index < other.Index
}
