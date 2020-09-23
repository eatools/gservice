package logmod

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

var (
	ServerLog1 *logrus.Logger
	logConf    *LogConfig
)

func InitLog(logc *LogConfig) {
	ServerLog1 = NewLog()
	logConf = logc
}

func Type(logType string) *logrus.Entry {
	return ServerLog1.WithField("logtype", logType)
}

func NewLog() *logrus.Logger {
	//logrus.NewEntry()
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.DebugLevel)
	// //hook := logrustash.New(conn, logrustash.DefaultFormatter(logrus.Fields{"type": typeName}))
	// hook, err := logrustash.NewAsyncHook(conf.Schema, conf.Address, typeName)
	// if err != nil {
	// 	panic(err)
	// }
	if logConf.DisablePrint {
		log.SetOutput(ioutil.Discard) // 关闭到终端的输出
	}
	// hook.ReconnectBaseDelay = time.Second
	// hook.ReconnectDelayMultiplier = 1
	// hook.MaxReconnectRetries = 3

	// log.AddHook(hook)

	gzipHook := NewGzipHook(logConf.TempPath, logConf.SplitMaxSize)
	log.AddHook(gzipHook)
	return log
}
