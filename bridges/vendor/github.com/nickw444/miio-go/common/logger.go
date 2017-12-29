package common

import "github.com/sirupsen/logrus"

var (
	Log *logrus.Logger = logrus.New()
)

func init() {
	Log.SetLevel(logrus.WarnLevel)
}

func SetLogger(logger *logrus.Logger) {
	Log = logger
}
