# [connsvr](http://github.com/simplejia/connsvr) 长连接推送服务
## 功能
* 支持tcp自定义协议长连接
* 支持http协议长连接
* 每个用户建立一个连接，每个连接唯一对应一个用户，用户可以同时加入多个房间
* 推送数据时，可以不给房间内特定用户推数据
* 接收到上行数据后，同步转发给相应业务处理服务
* 支持定期拉取远程服务器配置数据
* 根据远程配置信息，可给客户端下发消息拉取方式指令，目前支持:
  * 推送通知，然后客户端主动拉后端服务  
  * 推送整条消息，客户端不用拉
  * 推送通知，然后客户端来connsvr拉消息(connsvr为支持此功能，需进一步开发，预计很快更新代码)

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
http://xxx.xxx.com/enter?rid=xxx&uid=xxx&sid=xxx&callback=xxx
请求参数说明:
rid: 房间号
uid: 用户id
sid: session_id，区分同一uid不同连接，[可选]
callback: jsonp回调函数，[可选]

返回数据说明：
[callback(][json body][)]
示例如下: cb({"body":"hello world","cmd":"99","rid":"r1","sid":"","subcmd":"0","uid":"r2"})
```

* tcp自定义协议长连接（包括收包，回包）
```
Sbyte+Length+Cmd+Subcmd+UidLen+Uid+SidLen+Sid+RidLen+Rid+BodyLen+Body+ExtLen+Ext+Ebyte

Sbyte: 1个字节，固定值：0xfa，标识数据包开始
Length: 2个字节(网络字节序)，包括自身在内整个数据包的长度
Cmd: 1个字节，0x01：心跳 0x02：加入房间 0x03：退出房间 0x04：上行消息 0xff：connsvr异常
Subcmd: 1个字节，当用于上行消息时，路由不同的后端接口
UidLen: 1个字节，代表Uid长度
Uid: 用户id，对于app，可以是设备id，对于浏览器，可以是登陆用户id
SidLen: 1个字节，代表Sid长度
Sid: session_id，区分同一uid不同连接，[可选]对于浏览器，可以是生成的随机串，浏览器多窗口，多标签需单独生成随机串
RidLen: 1个字节，代表Rid长度
Rid: 房间id
BodyLen: 2个字节(网络字节序)，代表Body长度
Body: 和业务方对接，connsvr会中转给业务方，中转给业务方数据示例如下：uid=u1&rid=r1&cmd=99&subcmd=0&body=hello，数据路由见conf/conf.json pubs节点
ExtLen: 2个字节(网络字节序)，代表Ext长度
Ext: 扩展字段，当来自于connsvr时，目前支持如下：
```
{
    "GetMsgKind": 1 // 1: 推送通知，然后客户端主动拉后端服务  2: 推送整条消息，客户端不用拉 3: 推送通知，然后客户端来connsvr拉消息
}
```
Ebyte: 1个字节，固定值：0xfb，标识数据包结束

注1：上行数据包长度，即Length大小，限制4096字节内，下行不限
注2：当connsvr服务处理异常，比如调用后端服务失败，返回给client的数据包，Cmd：0xff
```

* 后端push协议格式(udp)
```
Cmd+Subcmd+UidLen+Uid+SidLen+Sid+RidLen+Rid+BodyLen+Body:

Cmd: 1个字节，经由connsvr直接转发给client
Subcmd: 1个字节，经由connsvr直接转发给client
UidLen: 1个字节，代表Uid长度
Uid: 指定排除的用户uid
SidLen: 1个字节，代表Sid长度
Sid: 指定排除的用户session_id，当没有传入Sid时，只匹配uid
RidLen: 1个字节，代表Rid长度
Rid: 房间id
BodyLen: 2个字节(网络字节序)，代表Body长度
Body: 和业务方对接，connsvr会中转给client

注：数据包长度限制50k内
```

## 使用方法
* 配置文件：[conf.json](http://github.com/simplejia/connsvr/tree/master/conf/conf.json) (json格式，支持注释)，可以通过传入自定义的env及conf参数来重定义配置文件里的参数，如：./connsvr -env dev -conf='hport=80;clog.mode=1'，多个参数用`;`分隔
* 建议用[cmonitor](http://github.com/simplejia/cmonitor)做进程启动管理
* api文件夹提供的代码用于后端服务给connsvr推送消息的，实际是通过[clog](http://github.com/simplejia/clog)服务分发的
* connsvr的上报数据，比如本机ip定期上报（用于更新待推送服务器列表），连接数、推送用时上报，等等，这些均是通过clog服务中转实现，所以我提供了clog的handler，均在testdata目录里：相应要修改clog的conf.json部分如下：
```
"connsvr/logbusi_report": [
    {
        "handler": "connreporthandler",
        "params": {
            "redis": {"addrtype": "ip", "addr": ":6379"}
        }
    }
],
"connsvr/logbusi_stat": [
    {
        "handler": "connstathandler",
        "params": {}
    }
],
"demo/logbusi_push": [
    {
        "handler": "connpushhandler",
        "params": {
            "redis": {"addrtype": "ip", "addr": ":6379"}
        }
    }
]
```
