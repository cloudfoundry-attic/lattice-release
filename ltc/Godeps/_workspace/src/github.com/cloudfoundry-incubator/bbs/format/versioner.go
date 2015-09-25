package format

import "github.com/gogo/protobuf/proto"

type Version byte

const (
	V0 Version = 0
)

//go:generate counterfeiter . Versioner
type Versioner interface {
	proto.Message
	MigrateFromVersion(v Version) error
	Validate() error
	Version() Version
}
