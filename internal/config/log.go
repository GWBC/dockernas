package config

import (
	"log"

	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger() {
	if IsBasePathSet() == false {
		return
	}
	logPath := GetLogPath()
	log.SetOutput(&lumberjack.Logger{
		Filename:   logPath + "/server.log",
		MaxSize:    64, // megabytes
		MaxBackups: 32,
		MaxAge:     30, //days
	})
}
