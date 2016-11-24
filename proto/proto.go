package proto

import (
	"fmt"

	"github.com/simplejia/connsvr/comm"
)

type Msg interface {
	Length() int
	SetLength(int)
	Cmd() comm.CMD
	SetCmd(comm.CMD)
	Subcmd() byte
	SetSubcmd(byte)
	Uid() string
	SetUid(string)
	Sid() string
	SetSid(string)
	Rid() string
	SetRid(string)
	Body() string
	SetBody(string)
	Misc() interface{}
	SetMisc(interface{})
	Encode() ([]byte, bool)
	Decode([]byte) bool
}

type MsgComm struct {
	length int
	cmd    comm.CMD
	subcmd byte
	uid    string
	sid    string
	rid    string
	body   string
	misc   interface{}
}

func (msg *MsgComm) SetMisc(misc interface{}) {
	msg.misc = misc
}

func (msg *MsgComm) Misc() interface{} {
	return msg.misc
}

func (msg *MsgComm) SetLength(length int) {
	msg.length = length
}

func (msg *MsgComm) Length() int {
	return msg.length
}

func (msg *MsgComm) Cmd() comm.CMD {
	return msg.cmd
}

func (msg *MsgComm) SetCmd(cmd comm.CMD) {
	msg.cmd = cmd
}

func (msg *MsgComm) Subcmd() byte {
	return msg.subcmd
}

func (msg *MsgComm) SetSubcmd(subcmd byte) {
	msg.subcmd = subcmd
}

func (msg *MsgComm) Body() string {
	return msg.body
}

func (msg *MsgComm) SetBody(body string) {
	msg.body = body
}

func (msg *MsgComm) Uid() string {
	return msg.uid
}

func (msg *MsgComm) SetUid(uid string) {
	msg.uid = uid
}

func (msg *MsgComm) Sid() string {
	return msg.sid
}

func (msg *MsgComm) SetSid(sid string) {
	msg.sid = sid
}

func (msg *MsgComm) Rid() string {
	return msg.rid
}

func (msg *MsgComm) SetRid(rid string) {
	msg.rid = rid
}

func (msg *MsgComm) Encode() ([]byte, bool) {
	return nil, false
}

func (msg *MsgComm) Decode([]byte) bool {
	return false
}

func NewMsg(t comm.PROTO) Msg {
	switch t {
	case comm.TCP:
		return new(MsgTcp)
	case comm.HTTP:
		return new(MsgHttp)
	case comm.UDP:
		return new(MsgUdp)
	default:
		panic(fmt.Sprintf("NewMsg() not support proto: %v", t))
	}
}
