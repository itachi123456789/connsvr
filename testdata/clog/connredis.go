package procs

import "fmt"

type ConnRedisAddr struct {
	AddrType string
	Addr     string
}

// 请赋值成自己的根据addrType, addr返回ip:port的函数
var ConnRedisAddrFunc = func(addrType, addr string) (string, error) {
	switch addrType {
	case "ip":
		return addr, nil
	default:
		return "", fmt.Errorf("ConnRedisAddrFunc() unexpected addrType: %s", addrType)
	}
}
