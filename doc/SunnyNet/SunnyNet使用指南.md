# SunnyNet 使用指南

## 1. 项目介绍

SunnyNet 是一个功能强大的网络中间件，类似于 Fiddler，但提供了更丰富的二次开发能力。它是一个跨平台的网络分析组件，可用于 HTTP/HTTPS/WS/WSS/TCP/UDP 网络分析和数据拦截修改。

### 主要功能：

- 支持获取/修改 HTTP/HTTPS/WS/WSS/TCP/TLS-TCP/UDP 发送及返回数据
- 支持对 HTTP/HTTPS/WS/WSS 指定连接使用指定代理
- 支持对 HTTP/HTTPS/WS/WSS/TCP/TLS-TCP 链接重定向
- 支持 gzip, deflate, br, ZSTD 解码
- 支持 WS/WSS/TCP/TLS-TCP/UDP 主动发送数据
- 支持 HTTPS 证书管理和自定义

## 2. 安装方法

### 2.1 依赖项

- Go 1.20 或更高版本（如需支持 Win7 系统，请使用 Go 1.21 以下版本，如 Go 1.20.4）
- Windows 系统推荐使用 TDM-GCC 编译器

### 2.2 安装步骤

1. 克隆仓库：

```bash
git clone https://github.com/qtgolang/SunnyNet.git
cd SunnyNet
```

2. 安装依赖：

```bash
go mod download
```

3. 编译项目：

```bash
go build
```

## 3. 基本使用方法

SunnyNet 的基本使用流程包括：创建实例、配置回调函数、设置端口、启动服务。

### 3.1 创建 SunnyNet 实例

```go
import "github.com/qtgolang/SunnyNet/SunnyNet"

// 创建 SunnyNet 实例
sunny := SunnyNet.NewSunny()
```

### 3.2 设置回调函数

```go
// 设置回调函数，用于处理各种网络事件
sunny.SetGoCallback(HttpCallback, TcpCallback, WSCallback, UdpCallback)
```

### 3.3 设置端口并启动

```go
// 设置端口并启动
sunny.SetPort(2025).Start()

// 检查是否有错误
if sunny.Error != nil {
    panic(sunny.Error)
}

// 防止程序退出
select {}
```

## 4. 回调函数实现

### 4.1 HTTP 回调

HTTP 回调函数用于处理 HTTP/HTTPS 请求和响应：

```go
func HttpCallback(Conn SunnyNet.ConnHTTP) {
    switch Conn.Type() {
    case public.HttpSendRequest: // 发起请求
        fmt.Println("发起请求", Conn.URL())
        
        // 可以修改请求头
        // Conn.GetRequestHeader().Set("User-Agent", "Custom User Agent")
        
        // 可以修改请求体
        // Conn.SetRequestBody([]byte("new request body"))
        
        // 可以直接响应，不让请求发送到服务器
        // Conn.StopRequest(200, "Hello World")
        
    case public.HttpResponseOK: // 请求完成
        bs := Conn.GetResponseBody()
        fmt.Println("请求完成", Conn.URL(), len(bs), Conn.GetResponseHeader())
        
        // 可以修改响应体
        // Conn.SetResponseBody([]byte("modified response"))
        
    case public.HttpRequestFail: // 请求错误
        fmt.Println(time.Now(), Conn.URL(), Conn.Error())
    }
}
```

### 4.2 WebSocket 回调

WebSocket 回调函数用于处理 WebSocket 连接和消息：

```go
func WSCallback(Conn SunnyNet.ConnWebSocket) {
    switch Conn.Type() {
    case public.WebsocketConnectionOK: // 连接成功
        fmt.Println("PID", Conn.PID(), "Websocket 连接成功:", Conn.URL())
        
    case public.WebsocketUserSend: // 发送数据
        if Conn.MessageType() < 5 {
            fmt.Println("PID", Conn.PID(), "Websocket 发送数据:", Conn.MessageType(), "->", hex.EncodeToString(Conn.Body()))
        }
        
        // 可以修改发送的数据
        // Conn.SetBody([]byte("modified message"))
        
    case public.WebsocketServerSend: // 收到数据
        if Conn.MessageType() < 5 {
            fmt.Println("PID", Conn.PID(), "Websocket 收到数据:", Conn.MessageType(), "->", hex.EncodeToString(Conn.Body()))
        }
        
        // 可以修改收到的数据
        // Conn.SetBody([]byte("modified message"))
        
    case public.WebsocketDisconnect: // 连接关闭
        fmt.Println("PID", Conn.PID(), "Websocket 连接关闭", Conn.URL())
    }
}
```

### 4.3 TCP 回调

TCP 回调函数用于处理 TCP 连接和数据：

```go
func TcpCallback(Conn SunnyNet.ConnTCP) {
    switch Conn.Type() {
    case public.SunnyNetMsgTypeTCPAboutToConnect: // 即将连接
        mode := string(Conn.Body())
        fmt.Println("PID", Conn.PID(), "TCP 即将连接到:", mode, Conn.LocalAddress(), "->", Conn.RemoteAddress())
        
        // 可以修改目标连接地址
        // Conn.SetNewAddress("8.8.8.8:8080")
        
    case public.SunnyNetMsgTypeTCPConnectOK: // 连接成功
        fmt.Println("PID", Conn.PID(), "TCP 连接到:", Conn.LocalAddress(), "->", Conn.RemoteAddress(), "成功")
        
    case public.SunnyNetMsgTypeTCPClose: // 连接关闭
        fmt.Println("PID", Conn.PID(), "TCP 断开连接:", Conn.LocalAddress(), "->", Conn.RemoteAddress())
        
    case public.SunnyNetMsgTypeTCPClientSend: // 客户端发送数据
        fmt.Println("PID", Conn.PID(), "TCP 发送数据", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.Type(), Conn.BodyLen(), Conn.Body())
        
        // 可以修改发送的数据
        // Conn.SetBody([]byte("modified data"))
        
    case public.SunnyNetMsgTypeTCPClientReceive: // 客户端收到数据
        fmt.Println("PID", Conn.PID(), "收到数据", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.Type(), Conn.BodyLen(), Conn.Body())
        
        // 可以修改收到的数据
        // Conn.SetBody([]byte("modified data"))
    }
}
```

### 4.4 UDP 回调

UDP 回调函数用于处理 UDP 数据：

```go
func UdpCallback(Conn SunnyNet.ConnUDP) {
    switch Conn.Type() {
    case public.SunnyNetUDPTypeSend: // 客户端向服务器端发送数据
        fmt.Println("PID", Conn.PID(), "发送UDP", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.BodyLen())
        
        // 可以修改发送的数据
        // Conn.SetBody([]byte("modified data"))
        
    case public.SunnyNetUDPTypeReceive: // 服务器端向客户端发送数据
        fmt.Println("PID", Conn.PID(), "接收UDP", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.BodyLen())
        
        // 可以修改响应的数据
        // Conn.SetBody([]byte("modified data"))
        
    case public.SunnyNetUDPTypeClosed: // 关闭会话
        fmt.Println("PID", Conn.PID(), "关闭UDP", Conn.LocalAddress(), Conn.RemoteAddress())
    }
}
```

## 5. 高级功能

### 5.1 HTTPS 证书管理

SunnyNet 支持自定义 HTTPS 证书，用于 HTTPS 流量分析：

```go
// 创建证书管理器
cert := SunnyNet.NewCertManager()

// 加载 P12 证书
ok := cert.LoadP12Certificate("path/to/cert.p12", "password")
fmt.Println("载入P12:", ok)
fmt.Println("证书名称：", cert.GetCommonName())

// 将证书添加到 SunnyNet
sunny.AddHttpCertificate("domain.com", cert, SunnyNet.HTTPCertRules_Request)
```

### 5.2 代理设置

SunnyNet 支持设置全局代理或为特定请求设置代理：

```go
// 设置全局代理
sunny.SetGlobalProxy("socket://127.0.0.1:1080", 30000)

// 设置代理规则
sunny.CompileProxyRegexp("127.0.0.1;[::1];192.168.*")
```

### 5.3 TCP 规则设置

可以设置强制走 TCP 的规则：

```go
// 设置强制走 TCP 的规则
sunny.SetMustTcpRegexp("*.example.com", true)
```

### 5.4 HTTP 请求限制

可以设置 HTTP 请求的最大更新长度：

```go
// 设置 HTTP 请求最大更新长度
sunny.SetHTTPRequestMaxUpdateLength(10086)
```

## 6. 完整示例

以下是一个完整的 SunnyNet 使用示例：

```go
package main

import (
    "fmt"
    "github.com/qtgolang/SunnyNet/SunnyNet"
    "github.com/qtgolang/SunnyNet/src/encoding/hex"
    "github.com/qtgolang/SunnyNet/src/public"
    "log"
    "time"
)

func main() {
    // 创建 SunnyNet 实例
    s := SunnyNet.NewSunny()
    
    // 设置回调函数
    s.SetGoCallback(HttpCallback, TcpCallback, WSCallback, UdpCallback)
    
    // 设置代理规则
    s.CompileProxyRegexp("127.0.0.1;[::1];192.168.*")
    
    // 设置强制走 TCP 的规则
    s.SetMustTcpRegexp("example.com", true)
    
    // 设置端口并启动
    Port := 2025
    s.SetPort(Port).Start()
    
    // 设置 HTTP 请求最大更新长度
    s.SetHTTPRequestMaxUpdateLength(10086)
    
    // 检查是否有错误
    err := s.Error
    if err != nil {
        panic(err)
    }
    
    fmt.Println("SunnyNet 运行在端口:", Port)
    
    // 阻止程序退出
    select {}
}

func HttpCallback(Conn SunnyNet.ConnHTTP) {
    switch Conn.Type() {
    case public.HttpSendRequest: // 发起请求
        fmt.Println("发起请求", Conn.URL())
        return
    case public.HttpResponseOK: // 请求完成
        bs := Conn.GetResponseBody()
        log.Println("请求完成", Conn.URL(), len(bs), Conn.GetResponseHeader())
        return
    case public.HttpRequestFail: // 请求错误
        fmt.Println(time.Now(), Conn.URL(), Conn.Error())
        return
    }
}

func WSCallback(Conn SunnyNet.ConnWebSocket) {
    switch Conn.Type() {
    case public.WebsocketConnectionOK: // 连接成功
        log.Println("PID", Conn.PID(), "Websocket 连接成功:", Conn.URL())
        return
    case public.WebsocketUserSend: // 发送数据
        if Conn.MessageType() < 5 {
            log.Println("PID", Conn.PID(), "Websocket 发送数据:", Conn.MessageType(), "->", hex.EncodeToString(Conn.Body()))
        }
        return
    case public.WebsocketServerSend: // 收到数据
        if Conn.MessageType() < 5 {
            log.Println("PID", Conn.PID(), "Websocket 收到数据:", Conn.MessageType(), "->", hex.EncodeToString(Conn.Body()))
        }
        return
    case public.WebsocketDisconnect: // 连接关闭
        log.Println("PID", Conn.PID(), "Websocket 连接关闭", Conn.URL())
        return
    }
}

func TcpCallback(Conn SunnyNet.ConnTCP) {
    switch Conn.Type() {
    case public.SunnyNetMsgTypeTCPAboutToConnect: // 即将连接
        mode := string(Conn.Body())
        log.Println("PID", Conn.PID(), "TCP 即将连接到:", mode, Conn.LocalAddress(), "->", Conn.RemoteAddress())
        return
    case public.SunnyNetMsgTypeTCPConnectOK: // 连接成功
        log.Println("PID", Conn.PID(), "TCP 连接到:", Conn.LocalAddress(), "->", Conn.RemoteAddress(), "成功")
        return
    case public.SunnyNetMsgTypeTCPClose: // 连接关闭
        log.Println("PID", Conn.PID(), "TCP 断开连接:", Conn.LocalAddress(), "->", Conn.RemoteAddress())
        return
    case public.SunnyNetMsgTypeTCPClientSend: // 客户端发送数据
        log.Println("PID", Conn.PID(), "TCP 发送数据", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.Type(), Conn.BodyLen(), Conn.Body())
        return
    case public.SunnyNetMsgTypeTCPClientReceive: // 客户端收到数据
        log.Println("PID", Conn.PID(), "收到数据", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.Type(), Conn.BodyLen(), Conn.Body())
        return
    }
}

func UdpCallback(Conn SunnyNet.ConnUDP) {
    switch Conn.Type() {
    case public.SunnyNetUDPTypeSend: // 客户端向服务器端发送数据
        log.Println("PID", Conn.PID(), "发送UDP", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.BodyLen())
        return
    case public.SunnyNetUDPTypeReceive: // 服务器端向客户端发送数据
        log.Println("PID", Conn.PID(), "接收UDP", Conn.LocalAddress(), Conn.RemoteAddress(), Conn.BodyLen())
        return
    case public.SunnyNetUDPTypeClosed: // 关闭会话
        log.Println("PID", Conn.PID(), "关闭UDP", Conn.LocalAddress(), Conn.RemoteAddress())
        return
    }
}
```

## 7. 常见问题

### 7.1 证书问题

如果在使用 HTTPS 拦截时遇到证书问题，请确保：

1. 正确加载了证书
2. 客户端信任了 SunnyNet 的根证书
3. 证书的域名与请求的域名匹配

### 7.2 性能问题

在处理大量请求时，可能会遇到性能问题。建议：

1. 避免在回调函数中进行耗时操作
2. 对于大型数据传输，考虑使用异步处理
3. 适当增加 HTTP 请求最大更新长度

### 7.3 代理设置

如果代理设置不生效，请检查：

1. 代理地址格式是否正确
2. 代理服务器是否可访问
3. 代理规则是否正确配置

## 8. 联系方式

如有问题或需要帮助，可通过以下方式联系：

- QQ群: 751406884
- 二群: 545120699
- 网址: [https://esunny.vip/](https://esunny.vip/) 