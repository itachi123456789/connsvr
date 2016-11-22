# [connsvr](http://github.com/simplejia/connsvr) 长连接推送服务
## 功能
* 支持tcp自定义协议长连接
* 支持http协议长连接
* 每个用户建立一个连接，每个连接唯一对应一个用户，用户可以同时加入多个房间
* 推送数据时，可以不给房间内特定用户推数据
* 接收到上行数据后，同步转发给相应业务处理服务

## 实现
* 启用一个协程用于接收后端push数据，启用若干个协程用于管理房间用户，用户被hash到对应协程
* 每个协程需要通过管道接收数据，包括：加入房间，退出房间，推送消息
* 每个用户连接启一个读协程
* 无锁 

## 特点
* 通信协议足够简单高效
* 服务设计尽量简化，通用性好

## 协议
* http长连接
```
http://xxx.xxx.com/enter?rid=xxx&uid=xxx&callback=xxx
请求参数说明:
rid: 房间号
uid: 用户id
callback: jsonp回调函数，[可选]

返回数据说明：
[callback(][json body][)]
示例如下: cb({"body":"hello world","cmd":"99","rid":"r1","subcmd":"0","uid":"r2"})

注：不支持上行，http长连接上行可通过短连接实现
```

* tcp自定义协议长连接（包括收包，回包）
```
Sbyte+Length+Cmd+Subcmd+UidLen+Uid+RidLen+Rid+Body+Ebyte

Sbyte: 1个字节，固定值：0xfa，标识数据包开始
Length: 2个字节(网络字节序)，包括自身在内整个数据包的长度
Cmd: 1个字节，0x01：心跳 0x02：加入房间 0x03：退出房间 0x04：上行消息 0xff：connsvr异常
Subcmd: 1个字节，当用于上行消息时，路由不同的后端接口
UidLen: 1个字节，代表Uid长度
Uid: 用户id，对于app，可以是设备id，对于浏览器，可以是生成的随机串，浏览器多窗口，多标签需单独生成随机串
RidLen: 1个字节，代表Rid长度
Rid: 房间id
Body: 和业务方对接，connsvr会中转给业务方，中转给业务方数据示例如下：uid=u1&rid=r1&cmd=99&subcmd=0&body=hello，数据路由见conf/conf.json pubs节点
Ebyte: 1个字节，固定值：0xfb，标识数据包结束

注1：上行数据包长度，即Length大小，限制4096字节内，下行不限
注2：当connsvr服务处理异常，比如调用后端服务失败，返回给client的数据包，Cmd：0xff
注3：目前大部分常量限制都是在代码里定义的，后续也会考虑放在conf/conf.json里
```

* 后端push协议格式(udp)
```
Cmd+Subcmd+UidLen+Uid+RidLen+Rid+Body:

Cmd: 1个字节，经由connsvr直接转发给client
Subcmd: 1个字节，经由connsvr直接转发给client
UidLen: 1个字节，代表Uid长度
Uid: 指定排除的用户
RidLen: 1个字节，代表Rid长度
Rid: 房间id
Body: 和业务方对接，connsvr会中转给client

注：数据包长度限制50k内
```

## 使用方法
* 配置文件：[conf.json](http://github.com/simplejia/connsvr/tree/master/conf/conf.json) (json格式，支持注释)，可以通过传入自定义的env及conf参数来重定义配置文件里的参数，如：./connsvr -env dev -conf='hport=80;clog.mode=1'，多个参数用`;`分隔
* 建议用[cmonitor](http://github.com/simplejia/cmonitor)做进程启动管理
* api文件夹提供的代码用于后端服务给connsvr推送消息的，实际是通过[clog](http://github.com/simplejia/clog)服务分发的
* connsvr的上报数据，比如本机ip定期上报（用于更新待推送服务器列表），连接数上报，推送用时上报，等等，这些均是通过clog服务中转实现，所以我提供了clog的handler，近期我会更新到代码库里

