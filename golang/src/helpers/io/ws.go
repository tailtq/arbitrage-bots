package io

import (
	"log"
	"sync"

	"arbitrage-bot/helpers"

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
    helpers.Panic(err)

    return &WebSocketClient{
        Endpoint: endpoint,
        Conn: conn,
        Done: make(chan struct{}),
    }
}

func (wsc *WebSocketClient) Start(streamHandler func(data *[]byte)) {
    // Start a goroutine to read messages from the WebSocket, and call the streamHandler function
    go func() {
        defer wsc.Conn.Close()
        for {
            select {
            case <-wsc.Done:
                return
            default:
                _, dataByte, err := wsc.Conn.ReadMessage()
                helpers.Panic(err)
                streamHandler(&dataByte)
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
