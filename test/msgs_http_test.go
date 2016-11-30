package tests

import (
	"encoding/json"
	"fmt"
	"strconv"

	_ "github.com/simplejia/connsvr"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/utils"

	"net"
	"testing"
)

func TestMsgsHttp(t *testing.T) {
	cmd := 99
	rid := "r1"
	uid := "u_TestMsgsHttp"
	text := "hello world"
	msgId := ""

	func() {
		conn, err := net.Dial(
			"udp",
			fmt.Sprintf("%s:%d", utils.GetLocalIp(), conf.C.App.Bport),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		msg := proto.NewMsg(comm.UDP)
		msg.SetCmd(comm.CMD(cmd))
		msg.SetRid(rid)
		msg.SetUid(uid)
		msg.SetBody(text)
		msg.SetExt(`{"msgid": "1"}`)
		data, ok := msg.Encode()
		if !ok {
			t.Fatal("msg.Encode() error")
		}

		_, err = conn.Write(data)
		if err != nil {
			t.Fatal(err)
		}
	}()

	gpp := &utils.GPP{
		Uri: fmt.Sprintf("http://:%d/msgs", conf.C.App.Hport),
		Headers: map[string]string{
			"Connection": "Close",
		},
		Params: map[string]string{
			"rid":      rid,
			"mid":      msgId,
			"callback": "",
		},
	}
	resp, err := utils.Get(gpp)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]string
	json.Unmarshal(resp, &m)
	if _cmd := m["cmd"]; _cmd != strconv.Itoa(int(comm.MSGS)) {
		t.Errorf("get: %v, expected: %v", _cmd, comm.MSGS)
	}

	expect_body, _ := json.Marshal([]string{text})
	if body := m["body"]; body != string(expect_body) {
		t.Errorf("get: %s, expected: %s", body, expect_body)
	}
}
