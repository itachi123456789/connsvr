package tests

import (
	"fmt"

	_ "github.com/simplejia/connsvr"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/connsvr/cons"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/utils"

	"net"
	"sync"
	"testing"
	"time"
)

func TestTcp(t *testing.T) {
	cmd := 99
	rid := "r1"
	uid := "u1"
	text := "hello world"

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

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

		if int(_msg.Cmd()) != cmd {
			t.Errorf("get: %v, expected: %v", _msg.Cmd(), cmd)
		}
		if _msg.Uid() != uid {
			t.Errorf("get: %s, expected: %s", _msg.Uid(), uid)
		}
		if _msg.Rid() != rid {
			t.Errorf("get: %s, expected: %s", _msg.Rid(), rid)
		}
		if _msg.Body() != text {
			t.Errorf("get: %s, expected: %s", _msg.Body(), text)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond * 50)

		conn, err := net.Dial(
			"udp",
			fmt.Sprintf("%s:%d", utils.GetLocalIp(), conf.C.App.Bport),
		)
		if err != nil {
			t.Fatal(err)
		}

		msg := proto.NewMsg(cons.UDP)
		msg.SetCmd(cons.CMD(cmd))
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
	}()

	wg.Wait()
}
