// Package sockets provides a websocket-based messaging system.
package sockets

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"

// 	is "github.com/kubex-ecosystem/gdbase/internal/services"
// 	"github.com/gorilla/websocket"
// )

// type Publisher struct {
// 	conn   *websocket.Conn
// 	config is.WebSocketConfig
// }

// func NewPublisher(config is.WebSocketConfig) *Publisher {
// 	return &Publisher{
// 		config: config,
// 	}
// }

// func (p *Publisher) Connect() error {
// 	var err error
// 	p.conn, _, err = websocket.DefaultDialer.Dial(p.config.URL, nil)
// 	if err != nil {
// 		return fmt.Errorf("failed to connect to WebSocket: %w", err)
// 	}
// 	log.Printf("WebSocket Publisher connected to %s", p.config.URL)
// 	return nil
// }

// func (p *Publisher) PublishMessage(topic string, data interface{}) error {
// 	if p.conn == nil {
// 		if err := p.Connect(); err != nil {
// 			return err
// 		}
// 	}

// 	jsonData, err := json.Marshal(data)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal data: %w", err)
// 	}

// 	message := fmt.Sprintf("%s %s", topic, string(jsonData))
// 	err = p.conn.WriteMessage(websocket.TextMessage, []byte(message))
// 	if err != nil {
// 		return fmt.Errorf("failed to send WebSocket message: %w", err)
// 	}

// 	log.Printf("Published WebSocket message: %s", topic)
// 	return nil
// }

// func (p *Publisher) Close() error {
// 	if p.conn != nil {
// 		return p.conn.Close()
// 	}
// 	return nil
// }
