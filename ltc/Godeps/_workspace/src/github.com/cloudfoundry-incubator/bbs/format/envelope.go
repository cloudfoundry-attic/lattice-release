package format

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gogo/protobuf/proto"
	"github.com/pivotal-golang/lager"
)

type EnvelopeFormat byte

const (
	LEGACY_JSON EnvelopeFormat = 0
	JSON        EnvelopeFormat = 1
	PROTO       EnvelopeFormat = 2
)

const EnvelopeOffset int = 2

func UnmarshalEnvelope(logger lager.Logger, unencodedPayload []byte, model Versioner) error {
	envelopeFormat, version := EnvelopeMetadataFromPayload(unencodedPayload)

	var err error
	switch envelopeFormat {
	case LEGACY_JSON:
		err = UnmarshalJSON(logger, unencodedPayload, model)
	case JSON:
		err = UnmarshalJSON(logger, unencodedPayload[EnvelopeOffset:], model)
	case PROTO:
		err = UnmarshalProto(logger, unencodedPayload[EnvelopeOffset:], model)
	default:
		err = fmt.Errorf("unknown format %s", envelopeFormat)
		logger.Error("cannot-unmarshal-unknown-serialization-format", err)
	}

	if err != nil {
		return err
	}

	model.MigrateFromVersion(version)

	err = model.Validate()
	if err != nil {
		logger.Error("invalid-record", err)
		return err
	}
	return nil
}

func MarshalEnvelope(format EnvelopeFormat, model Versioner) ([]byte, error) {
	var payload []byte
	var err error

	switch format {
	case PROTO:
		payload, err = MarshalProto(model)
	case JSON:
		payload, err = MarshalJSON(model)
	case LEGACY_JSON:
		return MarshalJSON(model)
	default:
		err = fmt.Errorf("unknown format %s", format)
	}

	if err != nil {
		return nil, err
	}

	// to avoid the following copy, change toProto to write the payload
	// into a buffer pre-filled with format and version.
	data := make([]byte, len(payload)+2)
	data[0] = byte(format)
	data[1] = byte(model.Version())
	for i := range payload {
		data[i+2] = payload[i]
	}

	return data, nil
}

func EnvelopeMetadataFromPayload(unencodedPayload []byte) (EnvelopeFormat, Version) {
	if !IsEnveloped(unencodedPayload) {
		return LEGACY_JSON, V0
	}
	return EnvelopeFormat(unencodedPayload[0]), Version(unencodedPayload[1])
}

func IsEnveloped(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	switch EnvelopeFormat(data[0]) {
	case JSON, PROTO:
	default:
		return false
	}

	switch Version(data[1]) {
	case V0:
	default:
		return false
	}

	return true
}

func UnmarshalJSON(logger lager.Logger, marshaledPayload []byte, model Versioner) error {
	err := json.Unmarshal(marshaledPayload, model)
	if err != nil {
		logger.Error("failed-to-json-unmarshal-payload", err)
		return err
	}
	return nil
}

func MarshalJSON(v Versioner) ([]byte, error) {
	if !isNil(v) {
		if err := v.Validate(); err != nil {
			return nil, err
		}
	}

	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func UnmarshalProto(logger lager.Logger, marshaledPayload []byte, model Versioner) error {
	err := proto.Unmarshal(marshaledPayload, model)
	if err != nil {
		logger.Error("failed-to-proto-unmarshal-payload", err)
		return err
	}
	return nil
}

func MarshalProto(v Versioner) ([]byte, error) {
	if !isNil(v) {
		if err := v.Validate(); err != nil {
			return nil, err
		}
	}

	bytes, err := proto.Marshal(v)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func isNil(a interface{}) bool {
	if a == nil {
		return true
	}

	switch reflect.TypeOf(a).Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return reflect.ValueOf(a).IsNil()
	}

	return false
}
