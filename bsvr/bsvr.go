package bsvr

import (
	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/comm"
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

	msg := proto.NewMsg(comm.UDP)
	request := make([]byte, 1024*50)
	for {
		readLen, err := conn.Read(request)
		if err != nil || readLen <= 0 {
			continue
		}

		ok := msg.Decode(request[:readLen])
		clog.Debug("Bserver() msg.Decode %+v, %v", msg, ok)
		if !ok {
			clog.Error("Bserver:Decode() %v", request[:readLen])
			continue
		}

		dispatchCmd(msg)
	}
}

func dispatchCmd(msg proto.Msg) {
	room.RM.Push(msg)
}
