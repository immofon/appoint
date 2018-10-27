package log

import (
	"fmt"
	"runtime"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/immofon/appoint/utils"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logger.Level = logrus.DebugLevel
	logger.Formatter = new(logrus.JSONFormatter)
	logger.ExitFunc = func(code int) {
		fmt.Println("exit", code)
		utils.Exit()
	}
}

func TextMode() {
	logger.SetFormatter(new(logrus.TextFormatter))
}

func L() *logrus.Entry {
	pc, file, line, _ := runtime.Caller(1)
	return logger.WithTime(time.Now()).
		WithField("$file", file).
		WithField("$line", line).
		WithField("$func", runtime.FuncForPC(pc).Name())
}

func E(err error) *logrus.Entry {
	return L().WithField("$err", err)
}
