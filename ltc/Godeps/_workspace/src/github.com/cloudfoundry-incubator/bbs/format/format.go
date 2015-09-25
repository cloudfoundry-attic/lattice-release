package format

import (
	"github.com/pivotal-golang/lager"
)

type Format struct {
	Encoding
	EnvelopeFormat
}

var (
	LEGACY_FORMATTING *Format = NewFormat(LEGACY_UNENCODED, LEGACY_JSON)
	FORMATTED_JSON    *Format = NewFormat(UNENCODED, JSON)
	ENCODED_PROTO     *Format = NewFormat(BASE64, PROTO)
)

func NewFormat(encoding Encoding, format EnvelopeFormat) *Format {
	return &Format{encoding, format}
}

func Marshal(format *Format, model Versioner) ([]byte, error) {
	envelopedPayload, err := MarshalEnvelope(format.EnvelopeFormat, model)
	if err != nil {
		return nil, err
	}
	return Encode(format.Encoding, envelopedPayload)
}

func Unmarshal(logger lager.Logger, encodedPayload []byte, model Versioner) error {
	unencodedPayload, err := Decode(encodedPayload)
	if err != nil {
		return err
	}
	return UnmarshalEnvelope(logger, unencodedPayload, model)
}
