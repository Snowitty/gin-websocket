package main

import (
	"github.com/gin-gonic/gin"
	"github.com/snowitty/gin-websocket/ws"
)


func main(){
    gin.SetMode(gin.ReleaseMode)

    go ws.Manager.Start()
    r := gin.Default()
    r.GET("/ws", ws.WsHandler)
    r.GET("pong", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })

    r.Run(":8282")
}
