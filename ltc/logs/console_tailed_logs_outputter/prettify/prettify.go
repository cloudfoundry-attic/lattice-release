package prettify

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/chug"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry/noaa/events"
	"github.com/pivotal-golang/lager"
)

var colorLookup = map[string]string{
	"executor":     "\x1b[33m",
	"rep":          "\x1b[34m",
	"garden-linux": "\x1b[35m",
}

func Prettify(logMessage *events.LogMessage) string {
	entry := chug.ChugLogMessage(logMessage)

	// TODO: Or, do we use GetSourceType() for raw and Json source for pretty?
	color, ok := colorLookup[strings.Split(entry.LogMessage.GetSourceType(), ":")[0]]
	if !ok {
		color = "\x1b[0m" // TODO: use a constant for this
	}
	var components []string

	// TODO: "no color" code is 1 char less than color codes; if sourceType not found in
	// colorLookup, only pad to 20 chars
	components = append(components, fmt.Sprintf("%-21s", colors.Colorize(color, entry.LogMessage.GetSourceType())))
	components = append(components, fmt.Sprintf("%-9s", entry.LogMessage.GetSourceInstance()))

	var whichFunc func(chug.Entry) []string

	if entry.IsLager {
		whichFunc = prettyPrintLog
	} else {
		whichFunc = prettyPrintRaw
	}

	components = append(components, whichFunc(entry)...)
	return strings.Join(components, " ")
}

func prettyPrintLog(entry chug.Entry) []string {

	color, ok := colorLookup[strings.Split(entry.LogMessage.GetSourceType(), ":")[0]]
	if !ok {
		color = "\x1b[0m" // TODO: use a constant for this
	}

	level := ""
	switch entry.Log.LogLevel {
	case lager.INFO:
		level = colors.Colorize(color, "[INFO]")
	case lager.DEBUG:
		level = colors.Gray("[DEBUG]")
	case lager.ERROR:
		level = colors.Red("[ERROR]")
	case lager.FATAL:
		level = colors.Red("[FATAL]")
	}
	level = fmt.Sprintf("%-15s", level)

	var components []string
	components = append(components, level)

	timestamp := entry.Log.Timestamp.Format("01/02 15:04:05.00")
	components = append(components, fmt.Sprintf("%-17s", timestamp))
	components = append(components, fmt.Sprintf("%-14s", entry.Log.Session))
	components = append(components, entry.Log.Message)

	if len(entry.Log.Data) > 0 {
		dataJSON, _ := json.Marshal(entry.Log.Data)
		components = append(components, fmt.Sprintf("\n%s%s", strings.Repeat(" ", 63), string(dataJSON)))
	}

	return components
}

func prettyPrintRaw(entry chug.Entry) []string {
	var components []string
	components = append(components, strings.Repeat(" ", 6)) // loglevel
	timestamp := time.Unix(0, entry.LogMessage.GetTimestamp())
	components = append(components, fmt.Sprintf("%-17s", timestamp.Format("01/02 15:04:05.00")))
	components = append(components, strings.Repeat(" ", 14)) // sesh
	components = append(components, string(entry.Raw))

	return components
}
