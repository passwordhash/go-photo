package config

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

type CustomFormatter struct {
	TimestampFormat string
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var color string
	switch entry.Level {
	case logrus.DebugLevel:
		color = "\x1b[36m" // циановый
	case logrus.InfoLevel:
		color = "\x1b[34m" // синий
	case logrus.WarnLevel:
		color = "\x1b[33m" // жёлтый
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		color = "\x1b[31m" // красный
	default:
		color = "\x1b[0m" // сброс цвета
	}

	timestamp := entry.Time.Format(f.TimestampFormat)

	message := strings.ReplaceAll(entry.Message, "\n", " ")

	var fields []string
	for k, v := range entry.Data {
		fields = append(fields, fmt.Sprintf("%s=%v", k, v))
	}
	fieldsStr := ""
	if len(fields) > 0 {
		fieldsStr = " | " + strings.Join(fields, " ")
	}

	var b bytes.Buffer
	_, err := fmt.Fprintf(&b, "%s[%s] %s | %s%s\x1b[0m\n",
		color, entry.Level.String(), timestamp, message, fieldsStr)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
