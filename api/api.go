package api

import (
	"encoding/json"

	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/comm"
)

type PushMsg struct {
	Cmd    byte
	Subcmd byte
	Uid    string
	Sid    string
	Rid    string
	Body   string
	Ext    string
}

// Push用来给connsvr推送消息，复用clog的功能
func Push(msg *PushMsg) error {
	bs, _ := json.Marshal(msg)
	clog.Busi(comm.BUSI_PUSH, "%s", bs)
	return nil
}
