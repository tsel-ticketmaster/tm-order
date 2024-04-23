package applogger

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tsel-ticketmaster/tm-order/pkg/hook"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
)

var (
	logrusLogger   *logrus.Logger
	logrusSyncOnce sync.Once
)

func constructLogrus() *logrus.Logger {
	formatter := &logrus.JSONFormatter{
		TimestampFormat:   time.RFC3339,
		DisableTimestamp:  false,
		DisableHTMLEscape: true,
		DataKey:           "",
		FieldMap:          logrus.FieldMap{logrus.FieldKeyFunc: "caller", logrus.FieldKeyLevel: "severity", logrus.FieldKeyTime: "timestamp"},
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcname := s[len(s)-1]
			filename := fmt.Sprintf("%s:%d", f.File, f.Line)
			return funcname, filename
		},
		PrettyPrint: false,
	}

	logger := logrus.New()
	logger.SetReportCaller(true)
	logger.SetFormatter(formatter)
	logger.AddHook(otellogrus.NewHook(otellogrus.WithLevels(logrus.AllLevels...)))
	logger.AddHook(hook.NewTraceIDLoggerHook())
	logger.AddHook(hook.NewStdoutLoggerHook(logrus.New(), formatter))

	return logger
}

func GetLogrus() *logrus.Logger {
	logrusSyncOnce.Do(func() {
		logrusLogger = constructLogrus()
	})

	return logrusLogger
}
