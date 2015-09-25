package models

//go:generate bash -c "protoc --proto_path=$GOPATH/src:$GOPATH/src/github.com/gogo/protobuf/protobuf/:. --gogoslick_out=. *.proto"
const (
	maximumAnnotationLength = 10 * 1024
	maximumRouteLength      = 4 * 1024
)
