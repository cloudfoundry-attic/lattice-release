package models

//go:generate bash ../scripts/generate_protos.sh

const (
	maximumAnnotationLength = 10 * 1024
	maximumRouteLength      = 4 * 1024
)
