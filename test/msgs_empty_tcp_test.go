package tests

import (
	"fmt"

	_ "github.com/simplejia/connsvr"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/utils"

	"net"
	"testing"
)

func TestMsgsEmptyTcp(t *testing.T) {
	subcmd := byte(1)
	rid := "r1"
	uid := "u_TestMsgsEmptyTcp"

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
	msg.SetSubcmd(subcmd)
	msg.SetUid(uid)
	msg.SetRid(rid)
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

	if _msg.Rid() != rid {
		t.Errorf("get: %s, expected: %s", _msg.Rid(), rid)
	}

	t.Log("get resp:", _msg.Body())
}
