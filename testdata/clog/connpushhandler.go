package procs

import (
	"encoding/json"
	"log"
	"net"

	"github.com/garyburd/redigo/redis"
	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/api"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/proto"
)

func ConnPushHandler(cate, subcate, body string, params map[string]interface{}) {
	pushMsg := &api.PushMsg{}
	err := json.Unmarshal([]byte(body), pushMsg)
	if err != nil {
		clog.Error("ConnPushHandler() json.Unmarshal body: %s, error: %v", body, err)
		return
	}

	var connRedisAddr *ConnRedisAddr
	bs, _ := json.Marshal(params["redis"])
	json.Unmarshal(bs, &connRedisAddr)
	if connRedisAddr == nil {
		log.Printf("ConnReportHandler() params not right: %v\n", params)
		return
	}

	addr, err := ConnRedisAddrFunc(connRedisAddr.AddrType, connRedisAddr.Addr)
	if err != nil {
		clog.Error("ConnPushHandler() ConnRedisAddrFunc error: %v", err)
		return
	}

	c, err := redis.Dial("tcp", addr)
	if err != nil {
		clog.Error("ConnPushHandler() redis.Dial error: %v", err)
		return
	}

	ips, err := redis.Strings(c.Do("ZRANGE", "conn:ips", 0, -1))
	if err != nil {
		clog.Error("ConnPushHandler() redis get ips error: %v", err)
		return
	}

	clog.Info("ConnPushHandler() ips: %v, pushMsg: %+v", ips, pushMsg)

	for _, ipport := range ips {
		conn, err := net.Dial("udp", ipport)
		if err != nil {
			clog.Error("ConnPushHandler() dial ipport: %s, error: %v", ipport, err)
			continue
		}
		defer conn.Close()

		msg := proto.NewMsg(comm.UDP)
		msg.SetCmd(comm.CMD(pushMsg.Cmd))
		msg.SetSubcmd(pushMsg.Subcmd)
		msg.SetUid(pushMsg.Uid)
		msg.SetSid(pushMsg.Sid)
		msg.SetRid(pushMsg.Rid)
		msg.SetBody(pushMsg.Body)
		msg.SetExt(pushMsg.Ext)
		data, ok := msg.Encode()
		if !ok {
			clog.Error("ConnPushHandler() msg encode error, ipport: %s, msg: %+v", ipport, msg)
			continue
		}
		_, err = conn.Write(data)
		if err != nil {
			clog.Error("ConnPushHandler() conn.Write ipport: %s, error: %v", ipport, err)
			continue
		}
	}

	return
}

func init() {
	RegisterHandler("connpushhandler", ConnPushHandler)
}
