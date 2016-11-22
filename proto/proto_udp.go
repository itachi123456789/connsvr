package proto

import (
	"runtime/debug"

	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/cons"
)

type MsgUdp struct {
	MsgComm
}

func (msg *MsgUdp) Decode(data []byte) (ok bool) {
	defer func() {
		if err := recover(); err != nil {
			clog.Error("MsgUdp:Decode() recover err: %v, stack: %s", err, debug.Stack())
			ok = false
		}
	}()

	pos := 0
	msg.cmd = cons.CMD(data[pos])
	pos += 1
	msg.subcmd = data[pos]
	pos += 1
	uid_len := int(data[pos])
	pos += 1
	msg.uid = string(data[pos : pos+uid_len])
	pos += uid_len
	rid_len := int(data[pos])
	pos += 1
	msg.rid = string(data[pos : pos+rid_len])
	pos += rid_len
	msg.body = string(data[pos:])

	return true
}

func (msg *MsgUdp) Encode() ([]byte, bool) {
	data := []byte{}
	data = append(data, byte(msg.cmd))
	data = append(data, msg.subcmd)
	data = append(data, byte(len(msg.uid)))
	data = append(data, msg.uid...)
	data = append(data, byte(len(msg.rid)))
	data = append(data, msg.rid...)
	data = append(data, msg.body...)

	return data, true
}
