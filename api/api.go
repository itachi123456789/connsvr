package api

import (
	"encoding/json"

	"github.com/simplejia/clog"
)

type PushMsg struct {
	Cmd    byte
	Submcd byte
	Uid    string
	Rid    string
	Body   string
}

// Push用来给connsvr推送消息，复用clog的功能
func Push(msg *PushMsg) error {
	bs, _ := json.Marshal(msg)
	clog.Busi("push", "%s", bs)
}
