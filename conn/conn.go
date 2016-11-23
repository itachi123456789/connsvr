package conn

import (
	"bytes"
	"net"

	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/proto"
)

type ConnWrap struct {
	T        comm.PROTO  // 消息类型
	C        net.Conn    // socket
	Uid      string      // 用户
	Rids     []string    // 房间列表
	Misc     interface{} // 额外参数
	LeftData []byte      // 粘包产生的多读的数据
}

// return false will close the conn
func (connWrap *ConnWrap) Read() (proto.Msg, bool) {
	switch connWrap.T {
	case comm.TCP:
		msg := new(proto.MsgTcp)
		rawData := make([]byte, comm.BUF_SIZE) // 至少要大于sbyte+length大小
		rawRead := 0
		if len(connWrap.LeftData) > 0 {
			rawData = append(connWrap.LeftData, rawData...)
			rawRead = len(connWrap.LeftData)
			connWrap.LeftData = nil
		}
		for rawRead < 3 { // magic number 3 is length of sbyte+length
			readLen, err := comm.ReadTimeout(connWrap.C, rawData[rawRead:], comm.C_RTIMEOUT)
			if err != nil || readLen <= 0 {
				clog.Warn("ConnWrap:ReadTimeout() 1, %v, %v", readLen, err)
				return nil, false
			}
			rawRead += readLen
		}

		skipRead, ok := msg.DecodeHeader(rawData[:rawRead])
		if !ok {
			clog.Warn("ConnWrap:DecodeHeader() %v", rawData[:rawRead])
			connWrap.LeftData = rawData[skipRead:rawRead]
			return nil, true
		}

		if msg.Length() > rawRead {
			if msg.Length() > len(rawData) {
				rawData = append(rawData, make([]byte, msg.Length()-len(rawData))...)
			}
			for rawRead < msg.Length() {
				readLen, err := comm.ReadTimeout(connWrap.C, rawData[rawRead:], comm.C_RTIMEOUT)
				if err != nil || readLen <= 0 {
					clog.Warn("ConnWrap:ReadTimeout() 2, %v, %v", readLen, err)
					return nil, false
				}
				rawRead += readLen
			}
		} else if msg.Length() < rawRead {
			connWrap.LeftData = rawData[msg.Length():rawRead]
		}

		if !msg.Decode(rawData[:rawRead]) {
			clog.Warn("ConnWrap:Decode() %v", rawData[:rawRead])
			return nil, true
		}

		return msg, true
	case comm.HTTP:
		msg := new(proto.MsgHttp)
		rawData := make([]byte, comm.BUF_SIZE4HTTP) // BUF_SIZE一定要大于http header第一行的长度
		rawRead := 0
		for rawRead < len(rawData) {
			readLen, err := comm.ReadTimeout(connWrap.C, rawData[rawRead:], comm.C_RTIMEOUT)
			if err != nil || readLen <= 0 {
				clog.Warn("ConnWrap:ReadTimeout() %v, %v", readLen, err)
				return nil, false
			}
			rawRead += readLen

			// check if get total first line
			if pos := bytes.IndexByte(rawData[rawRead-readLen:rawRead], '\n'); pos > 0 {
				break
			}
		}

		if !msg.Decode(rawData[:rawRead]) {
			clog.Warn("ConnWrap:Decode() %v", rawData[:rawRead])
			return nil, true
		}

		return msg, true
	default:
		clog.Error("ConnWrap unexpected T: %v", connWrap.T)
		return nil, false
	}
}

// when return false, close the connection
func (connWrap *ConnWrap) Write(msg proto.Msg) bool {
	data, ok := msg.Encode()
	if !ok {
		return true
	}
	for wlen := 0; wlen < len(data); {
		_wlen, err := comm.WriteTimeout(connWrap.C, data[wlen:], comm.C_WTIMEOUT)
		if err != nil || _wlen <= 0 {
			return false
		}
		wlen += _wlen
	}
	return true
}

func (connWrap *ConnWrap) Close() {
	// net.Conn可以多次关闭
	connWrap.C.Close()
}
