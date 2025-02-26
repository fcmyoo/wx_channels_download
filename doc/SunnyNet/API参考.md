# SunnyNet API 参考

本文档提供了 SunnyNet 主要 API 函数的详细参考。

## 1. 核心 API

### 1.1 创建和管理 SunnyNet 实例

#### NewSunny

创建一个新的 SunnyNet 实例。

```go
func NewSunny() *Sunny
```

**返回值**：
- `*Sunny` - SunnyNet 实例指针

**示例**：
```go
sunny := SunnyNet.NewSunny()
```

#### SetPort

设置 SunnyNet 实例的监听端口。

```go
func (s *Sunny) SetPort(port int) *Sunny
```

**参数**：
- `port` - 监听端口号

**返回值**：
- `*Sunny` - SunnyNet 实例指针，用于链式调用

**示例**：
```go
sunny.SetPort(8888)
```

#### Start

启动 SunnyNet 实例。

```go
func (s *Sunny) Start() *Sunny
```

**返回值**：
- `*Sunny` - SunnyNet 实例指针，用于链式调用

**示例**：
```go
sunny.Start()
```

#### Close

关闭 SunnyNet 实例。

```go
func (s *Sunny) Close()
```

**示例**：
```go
sunny.Close()
```

### 1.2 回调函数设置

#### SetGoCallback

设置 Go 回调函数，用于处理各种网络事件。

```go
func (s *Sunny) SetGoCallback(httpCallback func(ConnHTTP), tcpCallback func(ConnTCP), wsCallback func(ConnWebSocket), udpCallback func(ConnUDP)) *Sunny
```

**参数**：
- `httpCallback` - HTTP/HTTPS 回调函数
- `tcpCallback` - TCP 回调函数
- `wsCallback` - WebSocket 回调函数
- `udpCallback` - UDP 回调函数

**返回值**：
- `*Sunny` - SunnyNet 实例指针，用于链式调用

**示例**：
```go
sunny.SetGoCallback(HttpCallback, TcpCallback, WSCallback, UdpCallback)
```

### 1.3 代理设置

#### SetGlobalProxy

设置全局代理。

```go
func (s *Sunny) SetGlobalProxy(proxyURL string, timeout int) *Sunny
```

**参数**：
- `proxyURL` - 代理 URL，格式如 "socket5://user:pass@127.0.0.1:1080" 或 "http://user:pass@127.0.0.1:8080"
- `timeout` - 超时时间（毫秒）

**返回值**：
- `*Sunny` - SunnyNet 实例指针，用于链式调用

**示例**：
```go
sunny.SetGlobalProxy("socket5://127.0.0.1:1080", 30000)
```

#### CompileProxyRegexp

设置代理规则。

```go
func (s *Sunny) CompileProxyRegexp(regexp string) error
```

**参数**：
- `regexp` - 代理规则，格式如 "127.0.0.1;[::1];192.168.*"

**返回值**：
- `error` - 错误信息，如果没有错误则为 nil

**示例**：
```go
sunny.CompileProxyRegexp("127.0.0.1;[::1];192.168.*")
```

### 1.4 TCP 设置

#### MustTcp

设置是否强制所有连接走 TCP。

```go
func (s *Sunny) MustTcp(open bool)
```

**参数**：
- `open` - 是否开启强制 TCP

**示例**：
```go
sunny.MustTcp(true)
```

#### SetMustTcpRegexp

设置强制走 TCP 的规则。

```go
func (s *Sunny) SetMustTcpRegexp(regexpList string, rules bool) error
```

**参数**：
- `regexpList` - 规则列表，格式如 "*.example.com"
- `rules` - 规则模式，true 表示规则内的走 TCP，false 表示规则外的走 TCP

**返回值**：
- `error` - 错误信息，如果没有错误则为 nil

**示例**：
```go
sunny.SetMustTcpRegexp("*.example.com", true)
```

### 1.5 HTTP 设置

#### SetHTTPRequestMaxUpdateLength

设置 HTTP 请求最大更新长度。

```go
func (s *Sunny) SetHTTPRequestMaxUpdateLength(length int64)
```

**参数**：
- `length` - 最大长度（字节）

**示例**：
```go
sunny.SetHTTPRequestMaxUpdateLength(10086)
```

## 2. 证书管理 API

### 2.1 创建和管理证书

#### NewCertManager

创建一个新的证书管理器。

```go
func NewCertManager() *CertManager
```

**返回值**：
- `*CertManager` - 证书管理器指针

**示例**：
```go
cert := SunnyNet.NewCertManager()
```

#### LoadP12Certificate

加载 P12 证书。

```go
func (c *CertManager) LoadP12Certificate(path, password string) bool
```

**参数**：
- `path` - 证书文件路径
- `password` - 证书密码

**返回值**：
- `bool` - 是否成功加载

**示例**：
```go
ok := cert.LoadP12Certificate("path/to/cert.p12", "password")
```

#### GetCommonName

获取证书通用名称。

```go
func (c *CertManager) GetCommonName() string
```

**返回值**：
- `string` - 证书通用名称

**示例**：
```go
name := cert.GetCommonName()
```

### 2.2 添加证书到 SunnyNet

#### AddHttpCertificate

添加 HTTP 证书。

```go
func (s *Sunny) AddHttpCertificate(domain string, cert *CertManager, rules int) *Sunny
```

**参数**：
- `domain` - 域名
- `cert` - 证书管理器
- `rules` - 规则类型，如 SunnyNet.HTTPCertRules_Request

**返回值**：
- `*Sunny` - SunnyNet 实例指针，用于链式调用

**示例**：
```go
sunny.AddHttpCertificate("example.com", cert, SunnyNet.HTTPCertRules_Request)
```

## 3. HTTP 回调接口

### 3.1 ConnHTTP 接口

#### Type

获取当前消息事件类型。

```go
func (h *httpConn) Type() int
```

**返回值**：
- `int` - 消息类型，如 public.HttpSendRequest, public.HttpResponseOK, public.HttpRequestFail

**示例**：
```go
if Conn.Type() == public.HttpSendRequest {
    // 处理请求
}
```

#### URL

获取请求 URL。

```go
func (h *httpConn) URL() string
```

**返回值**：
- `string` - 请求 URL

**示例**：
```go
url := Conn.URL()
```

#### Method

获取请求方法。

```go
func (h *httpConn) Method() string
```

**返回值**：
- `string` - 请求方法，如 "GET", "POST"

**示例**：
```go
method := Conn.Method()
```

#### GetRequestHeader

获取请求头。

```go
func (h *httpConn) GetRequestHeader() http.Header
```

**返回值**：
- `http.Header` - 请求头

**示例**：
```go
headers := Conn.GetRequestHeader()
```

#### GetRequestBody

获取请求体。

```go
func (h *httpConn) GetRequestBody() []byte
```

**返回值**：
- `[]byte` - 请求体

**示例**：
```go
body := Conn.GetRequestBody()
```

#### SetRequestBody

设置请求体。

```go
func (h *httpConn) SetRequestBody(body []byte)
```

**参数**：
- `body` - 新的请求体

**示例**：
```go
Conn.SetRequestBody([]byte("new request body"))
```

#### GetResponseCode

获取响应状态码。

```go
func (h *httpConn) GetResponseCode() int
```

**返回值**：
- `int` - 响应状态码

**示例**：
```go
statusCode := Conn.GetResponseCode()
```

#### GetResponseHeader

获取响应头。

```go
func (h *httpConn) GetResponseHeader() http.Header
```

**返回值**：
- `http.Header` - 响应头

**示例**：
```go
headers := Conn.GetResponseHeader()
```

#### GetResponseBody

获取响应体。

```go
func (h *httpConn) GetResponseBody() []byte
```

**返回值**：
- `[]byte` - 响应体

**示例**：
```go
body := Conn.GetResponseBody()
```

#### SetResponseBody

设置响应体。

```go
func (h *httpConn) SetResponseBody(body []byte)
```

**参数**：
- `body` - 新的响应体

**示例**：
```go
Conn.SetResponseBody([]byte("new response body"))
```

#### StopRequest

阻止请求，直接返回响应。

```go
func (h *httpConn) StopRequest(statusCode int, data any, header ...http.Header)
```

**参数**：
- `statusCode` - 响应状态码
- `data` - 响应数据，可以是 string 或 []byte
- `header` - 可选的响应头

**示例**：
```go
Conn.StopRequest(200, "Hello World")
```

#### Error

获取错误信息。

```go
func (h *httpConn) Error() string
```

**返回值**：
- `string` - 错误信息

**示例**：
```go
err := Conn.Error()
```

## 4. WebSocket 回调接口

### 4.1 ConnWebSocket 接口

#### Type

获取当前消息事件类型。

```go
func (w *wsConn) Type() int
```

**返回值**：
- `int` - 消息类型，如 public.WebsocketConnectionOK, public.WebsocketUserSend, public.WebsocketServerSend, public.WebsocketDisconnect

**示例**：
```go
if Conn.Type() == public.WebsocketConnectionOK {
    // 处理连接成功
}
```

#### URL

获取 WebSocket URL。

```go
func (w *wsConn) URL() string
```

**返回值**：
- `string` - WebSocket URL

**示例**：
```go
url := Conn.URL()
```

#### MessageType

获取消息类型。

```go
func (w *wsConn) MessageType() int
```

**返回值**：
- `int` - 消息类型，1 表示文本，2 表示二进制

**示例**：
```go
if Conn.MessageType() == 1 {
    // 处理文本消息
}
```

#### Body

获取消息内容。

```go
func (w *wsConn) Body() []byte
```

**返回值**：
- `[]byte` - 消息内容

**示例**：
```go
body := Conn.Body()
```

#### SetBody

设置消息内容。

```go
func (w *wsConn) SetBody(body []byte)
```

**参数**：
- `body` - 新的消息内容

**示例**：
```go
Conn.SetBody([]byte("new message"))
```

#### PID

获取进程 ID。

```go
func (w *wsConn) PID() int
```

**返回值**：
- `int` - 进程 ID

**示例**：
```go
pid := Conn.PID()
```

## 5. TCP 回调接口

### 5.1 ConnTCP 接口

#### Type

获取当前消息事件类型。

```go
func (t *tcpConn) Type() int
```

**返回值**：
- `int` - 消息类型，如 public.SunnyNetMsgTypeTCPAboutToConnect, public.SunnyNetMsgTypeTCPConnectOK, public.SunnyNetMsgTypeTCPClose, public.SunnyNetMsgTypeTCPClientSend, public.SunnyNetMsgTypeTCPClientReceive

**示例**：
```go
if Conn.Type() == public.SunnyNetMsgTypeTCPAboutToConnect {
    // 处理即将连接
}
```

#### LocalAddress

获取本地地址。

```go
func (t *tcpConn) LocalAddress() string
```

**返回值**：
- `string` - 本地地址

**示例**：
```go
localAddr := Conn.LocalAddress()
```

#### RemoteAddress

获取远程地址。

```go
func (t *tcpConn) RemoteAddress() string
```

**返回值**：
- `string` - 远程地址

**示例**：
```go
remoteAddr := Conn.RemoteAddress()
```

#### Body

获取消息内容。

```go
func (t *tcpConn) Body() []byte
```

**返回值**：
- `[]byte` - 消息内容

**示例**：
```go
body := Conn.Body()
```

#### SetBody

设置消息内容。

```go
func (t *tcpConn) SetBody(body []byte)
```

**参数**：
- `body` - 新的消息内容

**示例**：
```go
Conn.SetBody([]byte("new data"))
```

#### BodyLen

获取消息内容长度。

```go
func (t *tcpConn) BodyLen() int
```

**返回值**：
- `int` - 消息内容长度

**示例**：
```go
length := Conn.BodyLen()
```

#### SetNewAddress

设置新的目标连接地址。

```go
func (t *tcpConn) SetNewAddress(address string) bool
```

**参数**：
- `address` - 新的目标地址

**返回值**：
- `bool` - 是否设置成功

**示例**：
```go
Conn.SetNewAddress("8.8.8.8:8080")
```

#### PID

获取进程 ID。

```go
func (t *tcpConn) PID() int
```

**返回值**：
- `int` - 进程 ID

**示例**：
```go
pid := Conn.PID()
```

## 6. UDP 回调接口

### 6.1 ConnUDP 接口

#### Type

获取当前消息事件类型。

```go
func (u *udpConn) Type() int
```

**返回值**：
- `int` - 消息类型，如 public.SunnyNetUDPTypeSend, public.SunnyNetUDPTypeReceive, public.SunnyNetUDPTypeClosed

**示例**：
```go
if Conn.Type() == public.SunnyNetUDPTypeSend {
    // 处理发送数据
}
```

#### LocalAddress

获取本地地址。

```go
func (u *udpConn) LocalAddress() string
```

**返回值**：
- `string` - 本地地址

**示例**：
```go
localAddr := Conn.LocalAddress()
```

#### RemoteAddress

获取远程地址。

```go
func (u *udpConn) RemoteAddress() string
```

**返回值**：
- `string` - 远程地址

**示例**：
```go
remoteAddr := Conn.RemoteAddress()
```

#### Body

获取消息内容。

```go
func (u *udpConn) Body() []byte
```

**返回值**：
- `[]byte` - 消息内容

**示例**：
```go
body := Conn.Body()
```

#### SetBody

设置消息内容。

```go
func (u *udpConn) SetBody(body []byte)
```

**参数**：
- `body` - 新的消息内容

**示例**：
```go
Conn.SetBody([]byte("new data"))
```

#### BodyLen

获取消息内容长度。

```go
func (u *udpConn) BodyLen() int
```

**返回值**：
- `int` - 消息内容长度

**示例**：
```go
length := Conn.BodyLen()
```

#### PID

获取进程 ID。

```go
func (u *udpConn) PID() int
```

**返回值**：
- `int` - 进程 ID

**示例**：
```go
pid := Conn.PID()
``` 