## 依赖包
`
github.com/gin-gonic/gin
github.com/gorilla/websocket
`

## 运行服务端
`go  run  mian.go`

自己搭建nginx或apache等web服务，分别在两个窗口运行

http://localhost/client.html?uid=1&to_uid=2

http://localhost/client.html?uid=2&to_uid=1

这样就可以聊天了
