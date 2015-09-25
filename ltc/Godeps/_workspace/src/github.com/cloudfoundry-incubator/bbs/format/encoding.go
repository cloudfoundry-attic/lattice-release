package format

import (
	"encoding/base64"
	"fmt"
)

type Encoding [EncodingOffset]byte

var (
	LEGACY_UNENCODED Encoding = [2]byte{}
	UNENCODED        Encoding = [2]byte{'0', '0'}
	BASE64           Encoding = [2]byte{'0', '1'}
)

const EncodingOffset int = 2

func Encode(encoding Encoding, payload []byte) ([]byte, error) {
	switch encoding {
	case LEGACY_UNENCODED:
		return payload, nil
	case UNENCODED:
		return append(encoding[:], payload...), nil
	case BASE64:
		encoded, err := EncodeBase64(payload)
		if err != nil {
			return nil, err
		}
		return append(encoding[:], encoded...), nil
	default:
		return nil, fmt.Errorf("Unknown encoding: %v", encoding)
	}
}

func Decode(payload []byte) ([]byte, error) {
	encoding := EncodingFromPayload(payload)
	switch encoding {
	case LEGACY_UNENCODED:
		return payload, nil
	case UNENCODED:
		return payload[EncodingOffset:], nil
	case BASE64:
		return DecodeBase64(payload[EncodingOffset:])
	default:
		return nil, fmt.Errorf("Unknown encoding: %v", encoding)
	}
}

func EncodeBase64(unencodedPayload []byte) ([]byte, error) {
	encodedLen := base64.StdEncoding.EncodedLen(len(unencodedPayload))
	encodedPayload := make([]byte, encodedLen)
	base64.StdEncoding.Encode(encodedPayload, unencodedPayload)
	return encodedPayload, nil
}

func DecodeBase64(encodedPayload []byte) ([]byte, error) {
	decodedLen := base64.StdEncoding.DecodedLen(len(encodedPayload))
	decodedPayload := make([]byte, decodedLen)
	n, err := base64.StdEncoding.Decode(decodedPayload, encodedPayload)
	return decodedPayload[:n], err
}

func EncodingFromPayload(payload []byte) Encoding {
	if !IsEncoded(payload) {
		return LEGACY_UNENCODED
	}
	return Encoding{payload[0], payload[1]}
}

func IsEncoded(payload []byte) bool {
	if len(payload) < EncodingOffset {
		return false
	}

	if payload[0] < '0' || payload[0] > '9' {
		return false
	}

	return true
}
