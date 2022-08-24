package ztimer

import (
	myLogger "github.com/lidaqi001/plugins/lib/logger"
	"log"
)

var errLog *log.Logger

func init() {
	// 初始化日志
	errLog = myLogger.ErrLogger("[ztimer]")
}
