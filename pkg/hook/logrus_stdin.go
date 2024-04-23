package hook

import (
	"os"

	"github.com/sirupsen/logrus"
)

type stdoutLoggerHook struct {
	logger *logrus.Logger
}

func NewStdoutLoggerHook(logger *logrus.Logger, formatter logrus.Formatter) logrus.Hook {
	logger.Out = os.Stdout
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(formatter)
	logger.SetReportCaller(true)

	return &stdoutLoggerHook{logger: logger}
}

// Levels implements logrus.Hook interface, this hook applies to all defined levels
func (d *stdoutLoggerHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.InfoLevel, logrus.DebugLevel}
}

// Fire implements logrus.Hook interface, attaches trace and span details found in entry context
func (d *stdoutLoggerHook) Fire(e *logrus.Entry) error {
	e.Logger = d.logger

	return nil
}
