package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ClientManager struct{
    Clinents map[string]*Client
    Broadcast  chan []byte
    Register   chan *Client
    Unregister chan *Client
}

type Client struct{
    ID  string
    Socket *websocket.Conn
    Send  chan []byte
}

type Message struct{
    Sender  string `json:"sender,omitempty"`
    Recipient string `json:"recipient,omitempty"`
    Content  string `json:"content,omitempty"`
}

var Manager = ClientManager{
    Broadcast: make(chan []byte),
    Register: make(chan *Client),
    Unregister: make(chan *Client),
    Clinents: make(map[string]*Client),
}

func (manager *ClientManager) Start(){
    for{
        log.Println("<---管道通信--->")
        select{
        case conn := <-Manager.Register:
            log.Printf("新用户加入: %v", conn.ID)
            Manager.Clinents[conn.ID] = conn
            jsonMessage,_ := json.Marshal(&Message{Content: "Sucessful connection to socket service"})
            conn.Send <- jsonMessage
        case conn := <-Manager.Unregister:
            log.Printf("用户离开: %v", conn.ID)
            if _, ok := Manager.Clinents[conn.ID]; ok{
                jsonMessage, _ := json.Marshal(&Message{Content:"A socket has disconnected"})
                conn.Send <- jsonMessage
                close(conn.Send)
                delete(Manager.Clinents, conn.ID)
            }
        case message := <-Manager.Broadcast:
            MessageStruct := Message{}
            json.Unmarshal(message, &MessageStruct)
            for id, conn := range Manager.Clinents{
                if id!=creatId(MessageStruct.Recipient, MessageStruct.Sender){
                    continue
                }
                select{
                case conn.Send <- message:
                default:
                    close(conn.Send)
                    delete(Manager.Clinents, conn.ID)
                }
            }
        }
    }
}

func creatId(uid, touid string) string{
    return uid+"_"+touid
}

func (c *Client) Read(){
    defer func() {
        Manager.Unregister <- c
        c.Socket.Close()
    }()

    for{
        c.Socket.PongHandler()
        _, message, err := c.Socket.ReadMessage()
        if err != nil{
            Manager.Unregister <- c
            c.Socket.Close()
            break
        }
        log.Printf("读取到客户端信息: %s", string(message))
        Manager.Broadcast <- message
    }
}

func (c *Client) Write(){
    defer func() {
        c.Socket.Close()
    }()

    for {
        select{
        case message, ok := <-c.Send:
            if !ok{
                c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            log.Printf("发送到客户端的消息: %s", string(message))


            c.Socket.WriteMessage(websocket.TextMessage, message)
        }
    }
} 

func WsHandler(c *gin.Context){
    uid := c.Query("uid")
    touid := c.Query("to_uid")
    conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {return true}}).Upgrade(c.Writer, c.Request, nil)
    if err != nil{
        http.NotFound(c.Writer, c.Request)
        return
    }

    client := &Client{
        ID:  creatId(uid, touid),
        Socket:  conn,
        Send: make(chan []byte),
    }

    Manager.Register <- client
    go client.Read()
    go client.Write()
}
