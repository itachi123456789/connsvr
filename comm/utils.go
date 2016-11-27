package comm

import (
	"net"
	"syscall"
	"time"
)

func GetRlimitFile() uint64 {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		panic(err)
	}
	return rLimit.Cur
}

func ReadTimeout(c net.Conn, data []byte, timeout string) (int, error) {
	dur, err := time.ParseDuration(timeout)
	if err != nil {
		return 0, err
	}
	c.SetReadDeadline(time.Now().Add(dur))
	return c.Read(data)
}

func WriteTimeout(c net.Conn, data []byte, timeout string) (int, error) {
	dur, err := time.ParseDuration(timeout)
	if err != nil {
		return 0, err
	}
	c.SetWriteDeadline(time.Now().Add(dur))
	return c.Write(data)
}
