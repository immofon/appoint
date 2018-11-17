package log

import (
	"fmt"
	"os"
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
	logger.SetOutput(os.Stdout)
}

func TextMode() {
	logger.SetFormatter(new(logrus.TextFormatter))
}

func l(skip int) *logrus.Entry {
	pc, file, line, _ := runtime.Caller(skip)
	return logger.WithTime(time.Now()).
		WithField("$file", file).
		WithField("$line", line).
		WithField("$func", runtime.FuncForPC(pc).Name())
}
func L() *logrus.Entry {
	return l(2)
}
func E(err error) *logrus.Entry {
	return l(2).WithField("$err", err)
}
