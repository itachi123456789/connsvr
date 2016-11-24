package api

import (
	"testing"

	"github.com/simplejia/clog"
)

func TestPush(t *testing.T) {
	clog.Init("connsvr", "", 14, 2)

	msg := &PushMsg{
		Cmd:    1,
		Subcmd: 2,
		Uid:    "u1",
		Sid:    "",
		Rid:    "r1",
		Body:   "text",
	}
	Push(msg)
}
