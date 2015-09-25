package models

type DomainSet map[string]struct{}

func (set DomainSet) Add(domain string) {
	set[domain] = struct{}{}
}

func (set DomainSet) Each(predicate func(domain string)) {
	for domain := range set {
		predicate(domain)
	}
}

func (set DomainSet) Contains(domain string) bool {
	_, found := set[domain]
	return found
}

func NewDomainSet(domains []string) DomainSet {
	domainSet := DomainSet{}
	for _, domain := range domains {
		domainSet.Add(domain)
	}
	return domainSet
}

func (request *UpsertDomainRequest) Validate() error {
	var validationError ValidationError

	if request.Domain == "" {
		return validationError.Append(ErrInvalidField{"domain"})
	}

	return nil
}
