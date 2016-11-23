// 长连接服务.
// author: simplejia
// date: 2015/11/19
package main

import (
	"fmt"
	"time"

	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/bsvr"
	_ "github.com/simplejia/connsvr/clog"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/connsvr/cons"
	"github.com/simplejia/connsvr/fsvr"
	"github.com/simplejia/utils"
)

func init() {
	// 定期上报，用于后端维护connsvr服务器列表
	go func() {
		tick := time.Tick(time.Minute)
		for {
			select {
			case <-tick:
				clog.Busi(cons.BUSI_REPORT, "%s", utils.GetLocalIp())
			}
		}
	}()
}

func main() {
	clog.Info("main() ulimit_nofile: %d", comm.GetRlimitFile())

	go fsvr.Fserver(fmt.Sprintf("%s:%d", "0.0.0.0", conf.C.App.Tport), cons.TCP)

	go fsvr.Fserver(fmt.Sprintf("%s:%d", "0.0.0.0", conf.C.App.Hport), cons.HTTP)

	go bsvr.Bserver(fmt.Sprintf("%s:%d", utils.GetLocalIp(), conf.C.App.Bport))

	select {}
}
