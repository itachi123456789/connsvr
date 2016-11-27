//go:generate stringer -type=CMD,PROTO -output=cons_string.go

package comm

import "time"

type CMD byte

const (
	PING CMD = iota + 1 // 1
	ENTER
	LEAVE
	PUB
	ERR = 0xff
)

type PROTO int

const (
	TCP PROTO = iota + 1 //1
	HTTP
	UDP
)

const (
	BUSI_REPORT = "report"
	BUSI_STAT   = "stat"
	BUSI_PUSH   = "push"
)

type Stat struct {
	Ip    string
	N     int
	Rid   string
	Msg   string
	Num   int
	Btime time.Time
	Etime time.Time
}
