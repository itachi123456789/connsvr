package tests

import (
	"fmt"

	_ "github.com/simplejia/connsvr"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/connsvr/cons"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/utils"

	"net"
	"testing"
)

func TestPub(t *testing.T) {
	rid := "r1"
	uid := "u1"
	text := "hello world"

	conn, err := net.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", utils.GetLocalIp(), conf.C.App.Tport),
	)
	if err != nil {
		t.Fatal(err)
	}

	msg := proto.NewMsg(cons.TCP)
	msg.SetCmd(cons.ENTER)
	msg.SetUid(uid)
	msg.SetRid(rid)
	msg.SetBody(text)
	data, ok := msg.Encode()
	if !ok {
		t.Fatal("msg.Encode() error")
	}

	_, err = conn.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	msg = proto.NewMsg(cons.TCP)
	msg.SetCmd(cons.PUB)
	msg.SetSubcmd(1)
	msg.SetUid(uid)
	msg.SetRid(rid)
	msg.SetBody(text)
	data, ok = msg.Encode()
	if !ok {
		t.Fatal("msg.Encode() error")
	}

	_, err = conn.Write(data)
	if err != nil {
		t.Fatal(err)
	}
}
