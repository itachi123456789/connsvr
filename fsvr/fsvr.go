package fsvr

import (
	"fmt"
	"net"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/connsvr/conn"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/connsvr/room"
	"github.com/simplejia/utils"
)

func Fserver(host string, t comm.PROTO) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		panic(err)
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		c, err := listener.AcceptTCP()
		if err != nil {
			clog.Error("Fserver() listener.AcceptTCP %v", err)
			continue
		}

		c.SetReadBuffer(conf.C.Cons.C_RBUF)
		c.SetWriteBuffer(conf.C.Cons.C_WBUF)

		connWrap := &conn.ConnWrap{
			T: t,
			C: c,
		}
		go frecv(connWrap)
	}
}

// 请赋值成自己的根据addrType, addr返回ip:port的函数
var PubAddrFunc = func(addrType, addr string) (string, error) {
	return addr, nil
}

func dispatchCmd(connWrap *conn.ConnWrap, msg proto.Msg) bool {
	switch msg.Cmd() {
	case comm.PING:
		return true
	case comm.ENTER:
		// 不同用户不能复用同一个连接, 新用户替代老用户数据
		if connWrap.Uid != msg.Uid() || connWrap.Sid != msg.Sid() {
			for _, rid := range connWrap.Rids {
				room.RM.Del(rid, connWrap)
			}
		}
		connWrap.Uid = msg.Uid()
		connWrap.Sid = msg.Sid()
		connWrap.Misc = msg.Misc()
		room.RM.Add(msg.Rid(), connWrap)
		return true
	case comm.LEAVE:
		room.RM.Del(msg.Rid(), connWrap)
		return true
	case comm.PUB:
		clog.Debug("dispatchCmd() comm.PUSH: %+v", msg)

		subcmd := strconv.Itoa(int(msg.Subcmd()))
		pub := conf.C.Pubs[subcmd]
		if pub == nil {
			clog.Error("dispatchCmd() no expected subcmd: %s", subcmd)
			return true
		}
		addr, err := PubAddrFunc(pub.AddrType, pub.Addr)
		if err != nil {
			clog.Error("dispatchCmd() PubAddrFunc error: %v", err)
			return true
		}
		arrs := []string{
			strconv.Itoa(int(msg.Cmd())),
			subcmd,
			msg.Uid(),
			msg.Sid(),
			msg.Rid(),
			url.QueryEscape(msg.Body()),
		}
		ps := map[string]string{}
		values, _ := url.ParseQuery(fmt.Sprintf(pub.Params, utils.Slice2Interface(arrs)...))
		for k, vs := range values {
			ps[k] = vs[0]
		}

		timeout, _ := time.ParseDuration(pub.Timeout)

		headers := map[string]string{
			"Host": pub.Host,
		}

		uri := fmt.Sprintf("http://%s/%s", addr, strings.TrimPrefix(pub.Cgi, "/"))

		gpp := &utils.GPP{
			Uri:     uri,
			Timeout: timeout,
			Headers: headers,
			Params:  ps,
		}

		var body []byte
		step, maxstep := -1, 3
		if pub.Retry < maxstep {
			maxstep = pub.Retry
		}
		for ; step < maxstep; step++ {
			switch pub.Method {
			case "get":
				body, err = utils.Get(gpp)
			case "post":
				body, err = utils.Post(gpp)
			}

			if err != nil {
				clog.Error("dispatchCmd() http error, err: %v, body: %s, gpp: %v, step: %d", err, body, gpp, step)
			} else {
				clog.Debug("dispatchCmd() http success, body: %s, gpp: %v", body, gpp)
				break
			}
		}

		if step == maxstep {
			msg.SetCmd(comm.ERR)
			msg.SetBody("")
		} else {
			msg.SetBody(string(body))
		}
		connWrap.Write(msg)
		return true
	case comm.MSGS:
		clog.Info("dispatchCmd() comm.MSGS: %+v", msg)

		msgId := msg.Body()
		bodys := room.ML.Bodys(msgId, msg)
		msg.SetBody(bodys)
		connWrap.Write(msg)
		return true
	default:
		clog.Warn("dispatchCmd() unexpected cmd: %v", msg.Cmd())
		return true
	}

	return true
}

func frecv(connWrap *conn.ConnWrap) {
	defer func() {
		if err := recover(); err != nil {
			clog.Error("frecv() recover err: %v, stack: %s", err, debug.Stack())
		}
		connWrap.Close()
		for _, rid := range connWrap.Rids {
			room.RM.Del(rid, connWrap)
		}
	}()

	for {
		msg, ok := connWrap.Read()
		clog.Debug("frecv() connWrap.Read %+v, %v", msg, ok)
		if !ok {
			return
		}

		if msg == nil {
			continue
		}

		ok = dispatchCmd(connWrap, msg)
		if !ok {
			return
		}
	}
}
