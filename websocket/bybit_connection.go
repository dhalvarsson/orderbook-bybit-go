package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"night/orderbook"

	"github.com/gorilla/websocket"
)

type BybitClient struct {
	Conn   *websocket.Conn
	Symbol string
	Depth  int
	done   chan struct{}
	ob     *orderbook.OrderBook
}

func InitBybitClient(symbol string, depth int) *BybitClient {
	return &BybitClient{
		Symbol: symbol,
		Depth:  depth,
		done:   make(chan struct{}),
		ob:     orderbook.InitBybitOrderBook(),
	}
}

func (c *BybitClient) Connect() error {
	url := url.URL{Scheme: "wss", Host: "stream.bybit.com", Path: "v5/public/linear"}
	log.Printf("Connection to %s", url.String())

	var err error
	c.Conn, _, err = websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return fmt.Errorf("connection error %v", err)

	}

	subMsg := map[string]interface{}{
		"op": "subscribe",
		"args": []string{
			fmt.Sprintf("orderbook.%d.%s", c.Depth, c.Symbol),
		},
	}

	if err := c.Conn.WriteJSON(subMsg); err != nil {
		return fmt.Errorf("subscription error %v", err)
	}

	return nil

}

func (c *BybitClient) readMessages() {

	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.Conn.ReadMessage()
			if err != nil {
				log.Println("read error", err)
			}

			//log.Printf("Orderbook messages: %s", message)

			c.handleMessage(message)
		}
	}
}

func (c *BybitClient) handleMessage(message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Println("error unmarshal message", err)
	}

	if msgType, ok := msg["type"].(string); ok {
		switch msgType {
		case "snapshot":
			var snapshot orderbook.OrderBookDepth
			if err := json.Unmarshal(message, &snapshot); err != nil {
				log.Println("error unmarshal snapshot", err)
			}
			c.ob.HandleSnapShot(&snapshot)

		case "delta":
			var delta orderbook.OrderBookDepth
			if err := json.Unmarshal(message, &delta); err != nil {
				log.Println("error unmarshal delta", err)
			}

			c.ob.HandleDelta(&delta)
		}
	}
}

func (c *BybitClient) Listen() {
	go c.readMessages()
}

func (c *BybitClient) Close() {
	close(c.done)
	c.Conn.Close()
}
