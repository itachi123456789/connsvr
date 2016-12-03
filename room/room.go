package room

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/connsvr/conn"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/lc"
	"github.com/simplejia/utils"
)

const (
	ADD = iota + 1 // 1
	DEL
	PUSH
)

type MsgElem struct {
	id   string
	body string
	uid  string
}

type MsgList []*MsgElem

func (a MsgList) Len() int           { return len(a) }
func (a MsgList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a MsgList) Less(i, j int) bool { return a[i].id < a[j].id }

func (a MsgList) Key4Lc(rid string) string {
	return "conn:msgs:" + rid
}

func (a MsgList) Append(id string, msg proto.Msg) {
	key_lc := a.Key4Lc(msg.Rid())
	a_lc, _ := lc.Get(key_lc)
	if a_lc != nil {
		a = a_lc.(MsgList)
	}

	x := &MsgElem{
		id:   id,
		body: msg.Body(),
		uid:  msg.Uid(),
	}
	i := sort.Search(len(a), func(i int) bool { return a[i].id >= id })
	if i == len(a) {
		a = append(a, x)
	} else if a[i].id == id {
		// unexpected here
	} else {
		a = append(a[:i], append([]*MsgElem{x}, a[i:]...)...)
	}

	n := conf.V.Get().ConnMsgNum
	if len(a) > n {
		a = a[len(a)-n:]
	}

	lc.Set(key_lc, a, time.Hour)
}

// 请赋值成自己的根据addrType, addr返回ip:port的函数
var MsgAddrFunc = func(addrType, addr string) (string, error) {
	return addr, nil
}

func (a MsgList) Bodys(id string, msg proto.Msg) (strs string) {
	key_lc := a.Key4Lc(msg.Rid())
	a_lc, ok := lc.Get(key_lc)
	if a_lc != nil {
		a = a_lc.(MsgList)
	}

	// 但connsvr缓存消息为空时，路由到后端服务拉取数据
	if len(a) == 0 && !ok {
		subcmd := strconv.Itoa(int(msg.Subcmd()))
		c := conf.C.Msgs[subcmd]
		if c == nil {
			clog.Error("MsgList:Bodys() no expected subcmd: %s", subcmd)
			return
		}
		addr, err := MsgAddrFunc(c.AddrType, c.Addr)
		if err != nil {
			clog.Error("MsgList:Bodys() MsgAddrFunc error: %v", err)
			return
		}
		arrs := []string{
			strconv.Itoa(int(msg.Cmd())),
			subcmd,
			msg.Uid(),
			msg.Sid(),
			msg.Rid(),
		}
		ps := map[string]string{}
		values, _ := url.ParseQuery(fmt.Sprintf(c.Params, utils.Slice2Interface(arrs)...))
		for k, vs := range values {
			ps[k] = vs[0]
		}

		timeout, _ := time.ParseDuration(c.Timeout)

		headers := map[string]string{
			"Host": c.Host,
		}

		uri := fmt.Sprintf("http://%s/%s", addr, strings.TrimPrefix(c.Cgi, "/"))

		gpp := &utils.GPP{
			Uri:     uri,
			Timeout: timeout,
			Headers: headers,
			Params:  ps,
		}

		body, err := utils.Get(gpp)
		if err != nil {
			clog.Error("MsgList:utils.Get() http error, err: %v, body: %s, gpp: %v", err, body, gpp)
			return
		}
		clog.Debug("MsgList:utils.Get() http success, body: %s, gpp: %v", body, gpp)

		var ms comm.Msgs
		err = json.Unmarshal(body, &ms)
		if err != nil {
			clog.Error("MsgList:json.Unmarshal() error, err: %v, body: %s, gpp: %v", err, body, gpp)
			return
		}

		for _, m := range ms {
			a = append(a, &MsgElem{
				id:   m.MsgId,
				body: m.Body,
				uid:  m.Uid,
			})
		}

		// 当后端也没有数据时，放一个空数据，避免下次再次拉取
		if len(a) == 0 {
			a = append(a, &MsgElem{})
		}

		lc.Set(key_lc, a, time.Hour)
	}

	i := sort.Search(len(a), func(i int) bool { return a[i].id > id })
	var bodys []string
	for _, e := range a[i:] {
		// 过滤掉自己的消息，但当客户端传入id为空时（客户端无缓存消息），不用过滤
		if id != "" {
			if e.uid == msg.Uid() {
				continue
			}
		}
		bodys = append(bodys, e.body)
	}
	bs, _ := json.Marshal(bodys)
	strs = string(bs)

	return
}

var ML MsgList

type roomMsg struct {
	cmd  int
	rid  string
	body interface{}
}

type RoomMap struct {
	n   int
	chs []chan *roomMsg
}

func (roomMap *RoomMap) init() {
	roomMap.n = conf.C.Cons.U_MAP_NUM
	roomMap.chs = make([]chan *roomMsg, roomMap.n)
	for i := 0; i < roomMap.n; i++ {
		roomMap.chs[i] = make(chan *roomMsg, 1e4)
		go roomMap.proc(i)
	}
}

func (roomMap *RoomMap) proc(i int) {
	ch := roomMap.chs[i]
	data := map[string]map[[2]string]*conn.ConnWrap{}

	for msg := range ch {
		switch msg.cmd {
		case ADD:
			rid := msg.rid
			connWrap := msg.body.(*conn.ConnWrap)
			rids_m, ok := data[rid]
			if !ok {
				rids_m = map[[2]string]*conn.ConnWrap{}
				data[rid] = rids_m
			}
			connWrap.Rids = append(connWrap.Rids, rid)
			for _, _rid := range connWrap.Rids[:len(connWrap.Rids)-1] {
				if _rid == rid {
					connWrap.Rids = connWrap.Rids[:len(connWrap.Rids)-1]
					break
				}
			}

			ukey := [2]string{connWrap.Uid, connWrap.Sid}
			rids_m[ukey] = connWrap
		case DEL:
			rid := msg.rid
			connWrap := msg.body.(*conn.ConnWrap)
			rids_m, ok := data[rid]
			if !ok {
				break
			}

			ukey := [2]string{connWrap.Uid, connWrap.Sid}
			delete(rids_m, ukey)
			if len(rids_m) == 0 {
				delete(data, rid)
			}
			for i, _rid := range connWrap.Rids {
				if _rid == rid {
					connWrap.Rids = append(connWrap.Rids[:i], connWrap.Rids[i+1:]...)
					break
				}
			}
		case PUSH:
			rid := msg.rid
			m := msg.body.(proto.Msg)
			rids_m, ok := data[rid]
			if !ok || len(rids_m) == 0 {
				break
			}

			ext := &comm.ServExt{
				GetMsgKind: conf.V.Get().GetMsgKind,
			}
			ext_bs, _ := json.Marshal(ext)

			btime := time.Now()
			ukey_ex := [2]string{m.Uid(), m.Sid()}
			for ukey, connWrap := range rids_m {
				_ukey := [2]string{connWrap.Uid, connWrap.Sid}
				if ukey != _ukey {
					connWrap.Close()
					delete(rids_m, ukey)
					continue
				}
				if ukey_ex[1] == "" { // 当后端没有传入sid时，只匹配uid
					if ukey_ex[0] == _ukey[0] {
						continue
					}
				} else {
					if ukey_ex == _ukey {
						continue
					}
				}
				msg := proto.NewMsg(connWrap.T)
				msg.SetCmd(m.Cmd())
				msg.SetSubcmd(m.Subcmd())
				msg.SetUid(connWrap.Uid)
				msg.SetSid(connWrap.Sid)
				msg.SetRid(m.Rid())
				if ext.GetMsgKind == comm.DISPLAY {
					msg.SetBody(m.Body())
				}
				msg.SetExt(string(ext_bs))
				msg.SetMisc(connWrap.Misc)
				if ok := connWrap.Write(msg); !ok {
					connWrap.Close()
					delete(rids_m, ukey)
					continue
				}
			}
			etime := time.Now()
			stat, _ := json.Marshal(&comm.Stat{
				Ip:    utils.LocalIp,
				N:     i,
				Rid:   rid,
				Msg:   fmt.Sprintf("%+v", m),
				Num:   len(rids_m),
				Btime: btime,
				Etime: etime,
			})
			clog.Busi(comm.BUSI_STAT, "%s", stat)
		default:
			clog.Error("RoomMap:proc() unexpected cmd %v", msg.cmd)
			return
		}
	}
}

func (roomMap *RoomMap) Add(rid string, connWrap *conn.ConnWrap) {
	if rid == "" || connWrap.Uid == "" {
		return
	}
	clog.Info("RoomMap:Add() %s, %+v", rid, connWrap)

	i := utils.Hash33(connWrap.Uid) % roomMap.n
	select {
	case roomMap.chs[i] <- &roomMsg{cmd: ADD, rid: rid, body: connWrap}:
	default:
		clog.Error("RoomMap:Add() chan full")
	}
}

func (roomMap *RoomMap) Del(rid string, connWrap *conn.ConnWrap) {
	if rid == "" || connWrap.Uid == "" {
		return
	}
	clog.Info("RoomMap:Del() %s, %+v", rid, connWrap)

	i := utils.Hash33(connWrap.Uid) % roomMap.n
	select {
	case roomMap.chs[i] <- &roomMsg{cmd: DEL, rid: rid, body: connWrap}:
	default:
		clog.Error("RoomMap:Del() chan full")
	}
}

func (roomMap *RoomMap) Push(msg proto.Msg) {
	clog.Info("RoomMap:Push() %+v", msg)

	var pushExt *comm.PushExt
	if ext := msg.Ext(); ext != "" {
		err := json.Unmarshal([]byte(ext), &pushExt)
		if err != nil {
			clog.Error("RoomMap:Push() json.Unmarshal error: %v", err)
			return
		}
		ML.Append(pushExt.MsgId, msg)
	}

	for _, ch := range roomMap.chs {
		select {
		case ch <- &roomMsg{cmd: PUSH, rid: msg.Rid(), body: msg}:
		default:
			clog.Error("RoomMap:Push() chan full")
		}
	}
}

var RM RoomMap

func init() {
	RM.init()
}
