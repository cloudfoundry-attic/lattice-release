package models

const (
	CellServiceName = "Cell"
)

type ServiceRegistration struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type ServiceRegistrations []ServiceRegistration

func (s ServiceRegistrations) FilterByName(serviceName string) ServiceRegistrations {
	registrations := ServiceRegistrations{}
	for _, reg := range s {
		if reg.Name == serviceName {
			registrations = append(registrations, reg)
		}
	}
	return registrations
}
