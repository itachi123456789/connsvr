package tests

import (
	"encoding/json"
	"fmt"

	_ "github.com/simplejia/connsvr"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/utils"

	"net"
	"testing"
)

func TestMsgsTcp(t *testing.T) {
	cmd := 99
	rid := "r1"
	uid := "u_TestMsgsTcp"
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

	conn, err := net.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", utils.GetLocalIp(), conf.C.App.Tport),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	msg := proto.NewMsg(comm.TCP)
	msg.SetCmd(comm.MSGS)
	msg.SetUid("")
	msg.SetRid(rid)
	msg.SetBody(msgId)
	data, ok := msg.Encode()
	if !ok {
		t.Fatal("msg.Encode() error")
	}

	_, err = conn.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	result := make([]byte, 4096)
	readLen, err := conn.Read(result)
	if err != nil || readLen <= 0 {
		t.Fatal(err, readLen)
	}

	_msg := new(proto.MsgTcp)
	_, ok = _msg.DecodeHeader(result[:readLen])
	if !ok {
		t.Fatal("_msg.DecodeHeader() error")
	}
	ok = _msg.Decode(result[:readLen])
	if !ok {
		t.Fatal("_msg.Decode() error")
	}

	if _msg.Cmd() == comm.ERR {
		t.Errorf("get: %v, expected: %v", _msg.Cmd(), msg.Cmd())
	}
	if _msg.Rid() != rid {
		t.Errorf("get: %s, expected: %s", _msg.Rid(), rid)
	}

	expect_body, _ := json.Marshal([]string{text})
	if body := _msg.Body(); body != string(expect_body) {
		t.Errorf("get: %s, expected: %s", _msg.Body(), expect_body)
	}
}
