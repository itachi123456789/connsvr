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

	"github.com/simplejia/utils"
)

type Conf struct {
	App *struct {
		Name  string
		Tport int
		Hport int
		Bport int
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

var (
	Envs map[string]*Conf
	Env  string
	C    *Conf
)

func init() {
	flag.StringVar(&Env, "env", "prod", "set env")
	var conf string
	flag.StringVar(&conf, "conf", "", "set custom conf")
	flag.Parse()

	dir := ""
	for _, p := range []string{".", ".."} {
		dir = filepath.Join(p, "conf")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			break
		}
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

	return
}
