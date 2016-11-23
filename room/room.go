package room

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/conn"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/utils"
)

const (
	ADD = iota + 1 // 1
	DEL
	PUSH
)

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
	roomMap.n = comm.U_MAP_NUM
	roomMap.chs = make([]chan *roomMsg, roomMap.n)
	for i := 0; i < roomMap.n; i++ {
		roomMap.chs[i] = make(chan *roomMsg, 1e5)
		go roomMap.proc(i)
	}
}

func (roomMap *RoomMap) proc(i int) {
	ch := roomMap.chs[i]
	data := map[string]map[string]*conn.ConnWrap{}

	for msg := range ch {
		switch msg.cmd {
		case ADD:
			rid := msg.rid
			connWrap := msg.body.(*conn.ConnWrap)
			rids_m, ok := data[rid]
			if !ok {
				rids_m = map[string]*conn.ConnWrap{}
				data[rid] = rids_m
			}
			connWrap.Rids = append(connWrap.Rids, rid)
			for _, _rid := range connWrap.Rids[:len(connWrap.Rids)-1] {
				if _rid == rid {
					connWrap.Rids = connWrap.Rids[:len(connWrap.Rids)-1]
					break
				}
			}
			rids_m[connWrap.Uid] = connWrap
		case DEL:
			rid := msg.rid
			connWrap := msg.body.(*conn.ConnWrap)
			rids_m, ok := data[rid]
			if !ok {
				break
			}
			delete(rids_m, connWrap.Uid)
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

			btime := time.Now()
			uid_ex := m.Uid()
			for uid, connWrap := range rids_m {
				if uid != connWrap.Uid {
					connWrap.Close()
					delete(rids_m, uid)
					continue
				}
				if uid_ex == connWrap.Uid {
					continue
				}
				msg := proto.NewMsg(connWrap.T)
				msg.SetCmd(m.Cmd())
				msg.SetSubcmd(m.Subcmd())
				msg.SetUid(connWrap.Uid)
				msg.SetRid(m.Rid())
				msg.SetBody(m.Body())
				msg.SetMisc(connWrap.Misc)
				if ok := connWrap.Write(msg); !ok {
					connWrap.Close()
					delete(rids_m, uid)
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

func (roomMap *RoomMap) Push(rid string, msg proto.Msg) {
	clog.Info("RoomMap:Push() %s, %+v", rid, msg)

	for _, ch := range roomMap.chs {
		select {
		case ch <- &roomMsg{cmd: PUSH, rid: rid, body: msg}:
		default:
			clog.Error("RoomMap:Push() chan full")
		}
	}
}

var RM RoomMap

func init() {
	RM.init()
}
