package logemitter

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/cloudfoundry/loggregatorlib/cfcomponent/generic_logger"
	"github.com/cloudfoundry/loggregatorlib/loggregatorclient"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

var (
	MAX_MESSAGE_BYTE_SIZE = (9 * 1024) - 512
	TRUNCATED_BYTES       = []byte("TRUNCATED")
	TRUNCATED_OFFSET      = MAX_MESSAGE_BYTE_SIZE - len(TRUNCATED_BYTES)
)

var ERR_SHARED_SECRET_NOT_SET = errors.New("Environment variable LOGGREGATOR_SHARED_SECRET is not set. Emitter requires a shared secret to sign log messages")

type Emitter interface {
	Emit(string, string)
	EmitError(string, string)
	EmitLogMessage(*events.LogMessage)
}

type LoggregatorEmitter struct {
	LoggregatorClient loggregatorclient.LoggregatorClient
	sn                string
	sId               string
	sharedSecret      string
	logger            generic_logger.GenericLogger
}

func isEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func splitMessage(message string) []string {
	return strings.FieldsFunc(message, func(r rune) bool {
		return r == '\n' || r == '\r'
	})
}

func (e *LoggregatorEmitter) Emit(appid, message string) {
	e.emit(appid, message, events.LogMessage_OUT)
}

func (e *LoggregatorEmitter) EmitError(appid, message string) {
	e.emit(appid, message, events.LogMessage_ERR)
}

func (e *LoggregatorEmitter) emit(appid, message string, messageType events.LogMessage_MessageType) {
	if isEmpty(appid) || isEmpty(message) {
		return
	}
	logMessage := e.newLogMessage(appid, message, messageType)
	e.logger.Debugf("Logging message from %s of type %s with appid %s and with data %s", *logMessage.SourceType, logMessage.MessageType, *logMessage.AppId, string(logMessage.Message))

	e.EmitLogMessage(logMessage)
}

func (e *LoggregatorEmitter) EmitLogMessage(logMessage *events.LogMessage) {
	messages := splitMessage(string(logMessage.GetMessage()))

	for _, message := range messages {
		if isEmpty(message) {
			continue
		}

		if len(message) > MAX_MESSAGE_BYTE_SIZE {
			logMessage.Message = append([]byte(message)[0:TRUNCATED_OFFSET], TRUNCATED_BYTES...)
		} else {
			logMessage.Message = []byte(message)
		}

		logEnvelope, err := e.newLogEnvelope(*logMessage.AppId, logMessage)
		if err != nil {
			e.logger.Errorf("Error creating envelope: %s", err)
			return
		}
		marshalledLogEnvelope, err := proto.Marshal(logEnvelope)
		if err != nil {
			e.logger.Errorf("Error marshalling envelope: %s", err)
			return
		}
		e.LoggregatorClient.Send(marshalledLogEnvelope)
	}
}

func NewEmitter(loggregatorServer, sourceName, sourceId string, debug bool) (*LoggregatorEmitter, error) {
	// TODO: delete when "legacy" format goes away
	sharedSecret := os.Getenv("LOGGREGATOR_SHARED_SECRET")
	if sharedSecret == "" {
		return nil, ERR_SHARED_SECRET_NOT_SET
	}

	e := &LoggregatorEmitter{sharedSecret: sharedSecret}

	e.sn = sourceName
	e.logger = generic_logger.NewDefaultGenericLogger(debug)
	e.LoggregatorClient = loggregatorclient.NewLoggregatorClient(loggregatorServer, e.logger, loggregatorclient.DefaultBufferSize)
	e.sId = sourceId

	e.logger.Debugf("Created new loggregator emitter: %#v", e)
	return e, nil
}

func (e *LoggregatorEmitter) newLogMessage(appId, message string, mt events.LogMessage_MessageType) *events.LogMessage {
	currentTime := time.Now()

	return &events.LogMessage{
		Message:        []byte(message),
		AppId:          proto.String(appId),
		MessageType:    &mt,
		SourceInstance: &e.sId,
		Timestamp:      proto.Int64(currentTime.UnixNano()),
		SourceType:     &e.sn,
	}
}

func (e *LoggregatorEmitter) newLogEnvelope(appId string, message *events.LogMessage) (*logmessage.LogEnvelope, error) {
	envelope := &logmessage.LogEnvelope{
		LogMessage: convertToOldFormat(message),
		RoutingKey: proto.String(appId),
		Signature:  []byte{},
	}
	err := envelope.SignEnvelope(e.sharedSecret)

	return envelope, err
}

func convertToOldFormat(message *events.LogMessage) *logmessage.LogMessage {
	return &logmessage.LogMessage{
		Message:     message.Message,
		AppId:       message.AppId,
		MessageType: logmessage.LogMessage_MessageType(message.GetMessageType()).Enum(),
		SourceName:  message.SourceType,
		SourceId:    message.SourceInstance,
		Timestamp:   message.Timestamp,
	}
}
