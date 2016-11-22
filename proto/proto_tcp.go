package proto

import (
	"encoding/binary"
	"runtime/debug"

	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/cons"
)

const (
	SBYTE = 0xfa
	EBYTE = 0xfb
)

type MsgTcp struct {
	MsgComm
}

func (msg *MsgTcp) DecodeHeader(data []byte) (skipRead int, ok bool) {
	pos := 0
	for ; pos < len(data); pos++ {
		if data[pos] == SBYTE {
			break
		}
	}
	if pos == len(data) {
		return len(data), false
	} else if pos > 0 {
		return pos, false
	}

	msg.length = int(binary.BigEndian.Uint16(data[1:3]))
	if msg.length > cons.BODY_LEN_LIMIT {
		return len(data), false
	}

	return 0, true
}

func (msg *MsgTcp) Decode(data []byte) (ok bool) {
	defer func() {
		if err := recover(); err != nil {
			clog.Error("MsgTcp:Decode() recover err: %v, stack: %s", err, debug.Stack())
			ok = false
		}
	}()

	pos := 0
	// skip sbyte+length
	pos += 3
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
	msg.body = string(data[pos : msg.length-1])
	pos = msg.length - 1
	ebyte := data[pos]
	if ebyte != EBYTE {
		return false
	}

	return true
}

func (msg *MsgTcp) Encode() ([]byte, bool) {
	data := []byte{}
	data = append(data, SBYTE)
	data = append(data, make([]byte, 2)...)
	data = append(data, byte(msg.cmd))
	data = append(data, msg.subcmd)
	data = append(data, byte(len(msg.uid)))
	data = append(data, msg.uid...)
	data = append(data, byte(len(msg.rid)))
	data = append(data, msg.rid...)
	data = append(data, msg.body...)
	data = append(data, EBYTE)
	binary.BigEndian.PutUint16(data[1:3], uint16(len(data)))

	return data, true
}
