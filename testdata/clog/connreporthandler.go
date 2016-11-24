package procs

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/simplejia/clog"
)

var connreportOnce sync.Once

func connreportTimer(connRedisAddr *ConnRedisAddr) {
	tick := time.Tick(time.Minute)
	for {
		select {
		case <-tick:
			addr, err := ConnRedisAddrFunc(connRedisAddr.AddrType, connRedisAddr.Addr)
			if err != nil {
				clog.Error("connreportTimer() ConnRedisAddrFunc error: %v", err)
				return
			}

			c, err := redis.Dial("tcp", addr)
			if err != nil {
				clog.Error("connreportTimer() redis.Dial error: %v", err)
				return
			}
			defer c.Close()

			min, max := 0, time.Now().Add(-1*time.Minute*30).Unix()
			c.Do("ZREMRANGEBYSCORE", "conn:report", min, max)
		}
	}
}

// body is a ip:port
func ConnReportHandler(cate, subcate, body string, params map[string]interface{}) {
	var connRedisAddr *ConnRedisAddr
	bs, _ := json.Marshal(params["redis"])
	json.Unmarshal(bs, &connRedisAddr)
	if connRedisAddr == nil {
		log.Printf("ConnReportHandler() params not right: %v\n", params)
		return
	}

	connreportOnce.Do(func() {
		go connreportTimer(connRedisAddr)
	})

	addr, err := ConnRedisAddrFunc(connRedisAddr.AddrType, connRedisAddr.Addr)
	if err != nil {
		clog.Error("ConnPushHandler() ConnRedisAddrFunc error: %v", err)
		return
	}

	c, err := redis.Dial("tcp", addr)
	if err != nil {
		clog.Error("ConnReportHandler() redis.Dial error: %v", err)
		return
	}
	defer c.Close()

	c.Do("ZADD", "conn:ips", time.Now().Unix(), body)
	return
}

func init() {
	RegisterHandler("connreporthandler", ConnReportHandler)
}
