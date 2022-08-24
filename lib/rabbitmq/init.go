package rabbitmq

import (
	myLogger "github.com/lidaqi001/plugins/lib/logger"
	"github.com/lidaqi001/plugins/lib/utils"
	"log"
)

var errLog *log.Logger

func init() {
	// 初始化日志
	errLog = myLogger.ErrLogger("[rabbitmq]")

	name, _ := utils.GetProgramName()
	myLogger.Init("./log/"+name, "debug")
}
