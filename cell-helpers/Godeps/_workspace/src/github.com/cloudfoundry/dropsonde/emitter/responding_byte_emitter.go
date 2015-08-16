package emitter

import "github.com/cloudfoundry/sonde-go/control"

type RespondingByteEmitter interface {
	ByteEmitter
	Respond(*control.ControlMessage)
}
