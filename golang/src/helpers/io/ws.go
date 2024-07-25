package io

import (
	"log"
	"sync"

	jsonHelper "arbitrage-bot/helpers/json"

	"github.com/gorilla/websocket"
)

type WebSocketClient struct {
    Endpoint string
    Conn *websocket.Conn
    Done chan struct{}
    StopOnce sync.Once
}

func NewWebSocketClient(endpoint string) *WebSocketClient {
    conn, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
    if err != nil {
        panic(err)
    }

    return &WebSocketClient{
        Endpoint: endpoint,
        Conn: conn,
        Done: make(chan struct{}),
    }
}

func (wsc *WebSocketClient) Start(streamHandler func(data interface{})) {
    // Start a goroutine to read messages from the WebSocket, and call the streamHandler function
    go func() {
        defer wsc.Conn.Close()
        for {
            select {
            case <-wsc.Done:
                return
            default:
                _, dataByte, err := wsc.Conn.ReadMessage()
                data := make(map[string]interface{})

                if err != nil {
                    log.Println("Error reading message:", err)
                    return
                }
                if err := jsonHelper.Unmarshal(dataByte, &data); err != nil {
                    panic(err)
                }
                streamHandler(data)
            }
        }
    }()
}

func (wsc *WebSocketClient) Stop() {
    wsc.StopOnce.Do(func() {
        close(wsc.Done)
        err := wsc.Conn.WriteMessage(
            websocket.CloseMessage,
            websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
        )
        if err != nil {
            log.Println("Error during closing websocket:", err)
        }
    })
}
