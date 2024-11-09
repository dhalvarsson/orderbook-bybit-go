package main

import (
	"log"
	"night/websocket"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	symbol := "ETHUSDT"
	depth := 500

	bybitConnection := websocket.InitBybitClient(symbol, depth)
	err := bybitConnection.Connect()
	if err != nil {
		log.Fatal("failed to connect to Bybit", err)
	}
	defer bybitConnection.Close()

	bybitConnection.Listen()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Shutting down....%s", sig)

}
