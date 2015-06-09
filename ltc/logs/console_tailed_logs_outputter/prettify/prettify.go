package prettify

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/logs/console_tailed_logs_outputter/chug"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/pivotal-golang/lager"
)

var colorLookup = map[string]string{
	"rep":          "\x1b[34m",
	"garden-linux": "\x1b[35m",
}

func Prettify(logMessage *events.LogMessage) string {
	entry := chug.ChugLogMessage(logMessage)

	// TODO: Or, do we use GetSourceType() for raw and Json source for pretty?
	color, ok := colorLookup[strings.Split(entry.LogMessage.GetSourceType(), ":")[0]]
	if !ok {
		color = colors.ColorDefault
	}

	sourcePrefix := fmt.Sprintf("[%s%s%s|%s%s%s]", color, entry.LogMessage.GetSourceType(), colors.ColorDefault, color, entry.LogMessage.GetSourceInstance(), colors.ColorDefault)
	colorWidth := len(color)*2 + len(colors.ColorDefault)*2
	sourcePrefixWidth := strconv.Itoa(22 + colorWidth)

	var components []string
	components = append(components, fmt.Sprintf("%-"+sourcePrefixWidth+"s", sourcePrefix))
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
	var logColor, level string
	switch entry.Log.LogLevel {
	case lager.INFO:
		logColor = colors.ColorDefault
		level = "[INFO]"
	case lager.DEBUG:
		logColor = colors.ColorGray
		level = "[DEBUG]"
	case lager.ERROR:
		logColor = colors.ColorRed
		level = "[ERROR]"
	case lager.FATAL:
		logColor = colors.ColorRed
		level = "[FATAL]"
	}
	level = fmt.Sprintf("%s%-9s", logColor, level)

	var components []string
	components = append(components, level)

	timestamp := entry.Log.Timestamp.Format("01/02 15:04:05.00")
	components = append(components, fmt.Sprintf("%-17s", timestamp))
	components = append(components, fmt.Sprintf("%-14s", entry.Log.Session))
	components = append(components, entry.Log.Message)
	components = append(components, colors.ColorDefault)

	if entry.Log.Error != nil {
		components = append(components, fmt.Sprintf("\n%s%s%s%s", strings.Repeat(" ", 66), logColor, entry.Log.Error.Error(), colors.ColorDefault))
	}

	if len(entry.Log.Data) > 0 {
		dataJSON, _ := json.Marshal(entry.Log.Data)
		components = append(components, fmt.Sprintf("\n%s%s%s%s", strings.Repeat(" ", 66), logColor, string(dataJSON), colors.ColorDefault))
	}

	return components
}

func prettyPrintRaw(entry chug.Entry) []string {
	var components []string
	components = append(components, strings.Repeat(" ", 9)) // loglevel
	timestamp := time.Unix(0, entry.LogMessage.GetTimestamp())
	components = append(components, fmt.Sprintf("%-17s", timestamp.Format("01/02 15:04:05.00")))
	components = append(components, strings.Repeat(" ", 14)) // sesh
	components = append(components, string(entry.Raw))

	return components
}
