package bsvr

import (
	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/cons"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/connsvr/room"

	"net"
)

func Bserver(host string) {
	udpAddr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	msg := proto.NewMsg(cons.UDP)
	request := make([]byte, 1024*50)
	for {
		readLen, err := conn.Read(request)
		if err != nil || readLen <= 0 {
			continue
		}
		if !msg.Decode(request[:readLen]) {
			clog.Error("Bserver:Decode() %v", request[:readLen])
			continue
		}

		clog.Info("Bserver() msg: %+v", msg)
		dispatchCmd(msg)
	}
}

func dispatchCmd(msg proto.Msg) {
	room.RM.Push(msg.Rid(), msg)
}
