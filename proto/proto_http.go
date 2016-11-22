package proto

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/simplejia/connsvr/cons"

	"fmt"
	"net/url"
)

type MsgHttp struct {
	MsgComm
}

func (msg *MsgHttp) Encode() ([]byte, bool) {
	data, _ := json.Marshal(map[string]string{
		"cmd":    strconv.Itoa(int(msg.cmd)),
		"subcmd": strconv.Itoa(int(msg.subcmd)),
		"uid":    msg.uid,
		"rid":    msg.rid,
		"body":   msg.body,
	})
	var resp []byte
	if callback, ok := msg.misc.(string); ok && callback != "" {
		resp = append(resp, callback...)
		resp = append(resp, '(')
		resp = append(resp, data...)
		resp = append(resp, ')')
	} else {
		resp = data
	}
	return []byte(
		fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
			"Content-Type:application/json;charset=UTF-8\r\n"+
			"Connection: Keep-Alive\r\n"+
			"Content-Length: %d\r\n\r\n%s",
			len(resp), resp,
		)), true
}

func (msg *MsgHttp) Decode(data []byte) bool {
	pos1 := bytes.IndexByte(data, ' ')
	if pos1 < 0 || pos1 >= len(data)-1 {
		return false
	}
	pos2 := bytes.IndexByte(data[pos1+1:], ' ')
	if pos2 < 0 {
		return false
	}
	pos2 += pos1 + 1
	rMethod, rUri := data[:pos1], data[pos1+1:pos2]
	if strings.ToUpper(string(rMethod)) != "GET" {
		return false
	}

	pUrl, err := url.ParseRequestURI(string(rUri))
	if err != nil {
		return false
	}

	switch pUrl.Path {
	case "/enter":
		values := pUrl.Query()
		rid, uid := values.Get("rid"), values.Get("uid")
		if rid == "" || uid == "" {
			return false
		}

		msg.rid = rid
		msg.uid = uid
		msg.misc = values.Get("callback")
		msg.cmd = cons.ENTER
		return true
	default:
		return false
	}
}
