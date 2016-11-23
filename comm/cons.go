//go:generate stringer -type=CMD,PROTO -output=cons_string.go

package comm

import "time"

const (
	BUF_SIZE       = 128                   // 一次读取数据大小, 大于大部分数据包长
	BUF_SIZE4HTTP  = 128                   // 一次读取数据大小, 大于http包头第一行
	BODY_LEN_LIMIT = 4096                  // 包最大长度
	U_MAP_NUM      = 300                   // 用户分组hash
	C_RTIMEOUT     = time.Minute * 2       // 读超时
	C_WTIMEOUT     = time.Millisecond * 10 // 写超时
)

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
