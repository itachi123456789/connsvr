package proto

import (
	"encoding/binary"
	"runtime/debug"

	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/comm"
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
	msg.cmd = comm.CMD(data[pos])
	pos += 1
	msg.subcmd = data[pos]
	pos += 1
	uid_len := int(data[pos])
	pos += 1
	msg.uid = string(data[pos : pos+uid_len])
	pos += uid_len
	sid_len := int(data[pos])
	pos += 1
	msg.sid = string(data[pos : pos+sid_len])
	pos += sid_len
	rid_len := int(data[pos])
	pos += 1
	msg.rid = string(data[pos : pos+rid_len])
	pos += rid_len
	body_len := int(binary.BigEndian.Uint16(data[pos : pos+2]))
	pos += 2
	msg.body = string(data[pos : body_len+pos])
	pos += body_len
	ext_len := int(binary.BigEndian.Uint16(data[pos : pos+2]))
	pos += 2
	msg.ext = string(data[pos : ext_len+pos])

	return true
}

func (msg *MsgUdp) Encode() ([]byte, bool) {
	data := []byte{}
	data = append(data, byte(msg.cmd))
	data = append(data, msg.subcmd)
	data = append(data, byte(len(msg.uid)))
	data = append(data, msg.uid...)
	data = append(data, byte(len(msg.sid)))
	data = append(data, msg.sid...)
	data = append(data, byte(len(msg.rid)))
	data = append(data, msg.rid...)
	data = append(data, make([]byte, 2)...)
	binary.BigEndian.PutUint16(data[len(data)-2:len(data)], uint16(len(msg.body)))
	data = append(data, msg.body...)
	data = append(data, make([]byte, 2)...)
	binary.BigEndian.PutUint16(data[len(data)-2:len(data)], uint16(len(msg.ext)))
	data = append(data, msg.ext...)

	return data, true
}
