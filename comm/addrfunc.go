package comm

func init() {
	// 请赋值成自己的根据addrType, addr返回ip:port的函数
	AddrFunc = func(addrType, addr string) (string, error) {
		switch addrType {
		case "zkname":
			/*
			   ip, port, err := zkname.GetHostByKey(addr)
			   if err != nil {
			       return "", err
			   }
			   return fmt.Sprintf("%s:%s", ip, port), nil
			*/
			return addr, nil
		default:
			return addr, nil
		}
	}
}
