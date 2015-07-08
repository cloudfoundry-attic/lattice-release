package chug

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/pivotal-golang/lager"
)

type Entry struct {
	IsLager    bool
	LogMessage *events.LogMessage
	Raw        []byte
	Log        LogEntry
}

type LogEntry struct {
	Timestamp time.Time
	LogLevel  lager.LogLevel
	Source    string
	Message   string
	Session   string
	Error     error
	Trace     string
	Data      lager.Data
}

func ChugLogMessage(logMessage *events.LogMessage) Entry {
	entry := Entry{
		IsLager:    false,
		LogMessage: logMessage,
		Raw:        logMessage.GetMessage(),
	}

	rawString := string(entry.Raw)
	idx := strings.Index(rawString, "{")
	if idx == -1 {
		return entry
	}

	var lagerLog lager.LogFormat
	decoder := json.NewDecoder(strings.NewReader(rawString[idx:]))
	err := decoder.Decode(&lagerLog)
	if err != nil {
		return entry
	}

	entry.Log, entry.IsLager = convertLagerLog(lagerLog)

	return entry
}

func convertLagerLog(lagerLog lager.LogFormat) (LogEntry, bool) {
	timestamp, err := strconv.ParseFloat(lagerLog.Timestamp, 64)
	if err != nil {
		return LogEntry{}, false
	}

	data := lagerLog.Data

	var logErr error
	if lagerLog.LogLevel == lager.ERROR || lagerLog.LogLevel == lager.FATAL {
		dataErr, ok := lagerLog.Data["error"]
		if ok {
			errorString, ok := dataErr.(string)
			if !ok {
				return LogEntry{}, false
			}
			logErr = errors.New(errorString)
			delete(lagerLog.Data, "error")
		}
	}

	var logTrace string
	dataTrace, ok := lagerLog.Data["trace"]
	if ok {
		logTrace, ok = dataTrace.(string)
		if !ok {
			return LogEntry{}, false
		}
		delete(lagerLog.Data, "trace")
	}

	var logSession string
	dataSession, ok := lagerLog.Data["session"]
	if ok {
		logSession, ok = dataSession.(string)
		if !ok {
			return LogEntry{}, false
		}
		delete(lagerLog.Data, "session")
	}

	return LogEntry{
		Timestamp: time.Unix(0, int64(timestamp*1e9)),
		LogLevel:  lagerLog.LogLevel,
		Source:    lagerLog.Source,
		Message:   lagerLog.Message,
		Session:   logSession,
		Error:     logErr,
		Trace:     logTrace,
		Data:      data,
	}, true
}
