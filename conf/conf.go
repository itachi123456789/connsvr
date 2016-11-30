package conf

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/utils"
)

type Conf struct {
	App *struct {
		Name  string
		Tport int
		Hport int
		Bport int
	}
	Cons *struct {
		BUF_SIZE       int    // 一次读取数据大小, 大于大部分数据包长
		BUF_SIZE4HTTP  int    // 一次读取数据大小, 大于http包头第一行
		BODY_LEN_LIMIT int    // 包最大长度
		U_MAP_NUM      int    // 用户分组hash
		C_RTIMEOUT     string // 读超时
		C_WTIMEOUT     string // 写超时
		C_RBUF         int    // 读缓冲区大小
		C_WBUF         int    // 写缓冲区大小
	}
	VarHost *struct {
		Addr     string
		AddrType string
		Host     string
		Cgi      string
	}
	Pubs map[string]*struct {
		Addr     string
		AddrType string
		Retry    int
		Host     string
		Cgi      string
		Params   string
		Method   string
		Timeout  string
	}
	Clog *struct {
		Name  string
		Mode  int
		Level int
	}
}

type Var struct {
	a          atomic.Value
	GetMsgKind comm.GET_MSG_KIND // 1: 推送通知，然后客户端主动拉后端服务  2: 推送整条消息，客户端不用拉 3: 推送通知，然后客户端来connsvr拉消息
	ConnMsgNum int               // connsvr服务缓存消息最大长度，在GetMsgKind参数为3时有效
}

func (v *Var) Get() *Var {
	return v.a.Load().(*Var)
}

func (v *Var) Set(nv *Var) {
	v.a.Store(nv)
}

var (
	Envs map[string]*Conf
	Env  string
	C    *Conf
	V    *Var
)

// 请赋值成自己的根据addrType, addr返回ip:port的函数
var VarAddrFunc = func(addrType, addr string) (string, error) {
	return addr, nil
}

func remoteConf() {
	V = &Var{
		GetMsgKind: comm.DISPLAY,
		ConnMsgNum: 20,
	}
	V.Set(V)

	go func() {
		for {
			time.Sleep(time.Minute)

			addr, err := VarAddrFunc(C.VarHost.AddrType, C.VarHost.Addr)
			if err != nil {
				clog.Error("remoteConf() VarAddrFunc error: %v", err)
				continue
			}
			headers := map[string]string{
				"Host": C.VarHost.Host,
			}
			uri := fmt.Sprintf("http://%s/%s", addr, strings.TrimPrefix(C.VarHost.Cgi, "/"))
			gpp := &utils.GPP{
				Uri:     uri,
				Headers: headers,
			}
			body, err := utils.Get(gpp)
			if err != nil {
				clog.Error("remoteConf() http error, err: %v, body: %s, gpp: %v", err, body, gpp)
				continue
			}

			v := &Var{}
			err = json.Unmarshal(body, v)
			if err != nil {
				clog.Error("remoteConf() json.Unmarshal error, err: %v, body: %s", err, body)
				continue
			}

			V.Set(v)
		}
	}()
}

func init() {
	flag.StringVar(&Env, "env", "prod", "set env")
	var conf string
	flag.StringVar(&conf, "conf", "", "set custom conf")
	flag.Parse()

	dir := "conf"
	for i := 0; i < 3; i++ {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			break
		}
		dir = filepath.Join("..", dir)
	}
	fcontent, err := ioutil.ReadFile(filepath.Join(dir, "conf.json"))
	if err != nil {
		panic(err)
	}

	fcontent = utils.RemoveAnnotation(fcontent)
	if err := json.Unmarshal(fcontent, &Envs); err != nil {
		panic(err)
	}

	C = Envs[Env]
	if C == nil {
		fmt.Println("env not right:", Env)
		os.Exit(-1)
	}

	func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("conf not right:", err)
				os.Exit(-1)
			}
		}()
		matchs := regexp.MustCompile(`[\w|\.]+|".*?[^\\"]"`).FindAllString(conf, -1)
		for n, match := range matchs {
			matchs[n] = strings.Replace(strings.Trim(match, "\""), `\"`, `"`, -1)
		}
		for n := 0; n < len(matchs); n += 2 {
			name, value := matchs[n], matchs[n+1]

			rv := reflect.Indirect(reflect.ValueOf(C))
			for _, field := range strings.Split(name, ".") {
				rv = reflect.Indirect(rv.FieldByName(strings.Title(field)))
			}
			switch rv.Kind() {
			case reflect.String:
				rv.SetString(value)
			case reflect.Bool:
				b, err := strconv.ParseBool(value)
				if err != nil {
					panic(err)
				}
				rv.SetBool(b)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				i, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					panic(err)
				}
				rv.SetInt(i)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				u, err := strconv.ParseUint(value, 10, 64)
				if err != nil {
					panic(err)
				}
				rv.SetUint(u)
			case reflect.Float32, reflect.Float64:
				f, err := strconv.ParseFloat(value, 64)
				if err != nil {
					panic(err)
				}
				rv.SetFloat(f)
			}
		}
	}()

	fmt.Printf("Env: %s\nC: %s\n", Env, utils.Iprint(C))

	remoteConf()

	return
}
