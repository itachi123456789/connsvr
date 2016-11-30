package room

import (
	"encoding/json"
	"fmt"
	"sort"
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

type msgElem struct {
	id   string
	body string
	uid  string
}

type msgList []*msgElem

func (a msgList) Len() int           { return len(a) }
func (a msgList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a msgList) Less(i, j int) bool { return a[i].id < a[j].id }

func (a msgList) Key4Lc(rid string) string {
	return "conn:msgs:" + rid
}

func (a msgList) Append(id, body, rid, uid string) {
	key_lc := a.Key4Lc(rid)
	a_lc, _ := lc.Get(key_lc)
	if a_lc != nil {
		a = a_lc.(msgList)
	}

	x := &msgElem{
		id:   id,
		body: body,
		uid:  uid,
	}
	i := sort.Search(len(a), func(i int) bool { return a[i].id >= id })
	if i == len(a) {
		a = append(a, x)
	} else if a[i].id == id {
		// unexpected here
	} else {
		a = append(a[:i], append([]*msgElem{x}, a[i:]...)...)
	}

	n := conf.V.Get().ConnMsgNum
	if len(a) > n {
		a = a[len(a)-n:]
	}

	lc.Set(key_lc, a, time.Hour)
}

func (a msgList) Bodys(id, rid, uid string) string {
	key_lc := a.Key4Lc(rid)
	a_lc, _ := lc.Get(key_lc)
	if a_lc != nil {
		a = a_lc.(msgList)
	}

	i := sort.Search(len(a), func(i int) bool { return a[i].id > id })
	var bodys []string
	for _, e := range a[i:] {
		// 过滤掉自己的消息
		if e.uid == uid {
			continue
		}
		bodys = append(bodys, e.body)
	}
	bs, _ := json.Marshal(bodys)
	return string(bs)
}

var ML msgList

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
		roomMap.chs[i] = make(chan *roomMsg, 1e5)
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
				Ip:    utils.GetLocalIp(),
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
		ML.Append(pushExt.MsgId, msg.Body(), msg.Rid(), msg.Uid())
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
