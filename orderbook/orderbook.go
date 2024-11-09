package orderbook

import (
	"errors"
	"log"
	"strconv"
	"sync"
	"time"
)

type OrderBook struct {
	lastUpdateId int64
	Bids         map[float64]float64
	Asks         map[float64]float64
	mu           sync.RWMutex
}

type OrderBookDepth struct {
	Topic string `json:"topic"`
	Data  struct {
		Symbol string     `json:"s"`
		Bids   [][]string `json:"b"`
		Asks   [][]string `json:"a"`
		Seq    int64      `json:"seq"`
	}
}

func InitBybitOrderBook() *OrderBook {
	return &OrderBook{
		Bids: make(map[float64]float64),
		Asks: make(map[float64]float64),
	}
}

func (ob *OrderBook) HandleSnapShot(snapshot *OrderBookDepth) error {
	start := time.Now()
	ob.mu.Lock()
	defer ob.mu.Unlock()

	ob.Bids = make(map[float64]float64)
	ob.Asks = make(map[float64]float64)
	ob.lastUpdateId = snapshot.Data.Seq

	ob.UpdateBook(snapshot.Data.Bids, snapshot.Data.Asks)

	elapsed := time.Since(start)
	log.Printf("Snapshot update took %s", &elapsed)

	return nil

}

func (ob *OrderBook) HandleDelta(delta *OrderBookDepth) error {
	start := time.Now()
	ob.mu.Lock()
	defer ob.mu.Unlock()

	ob.HandleLevels(delta.Data.Bids, ob.Bids, false)
	ob.HandleLevels(delta.Data.Asks, ob.Asks, false)
	ob.lastUpdateId = delta.Data.Seq

	/*
			if err := ob.LogMidPrice(); err != nil {
			log.Println("error calculation of midprice", err)
		}
	*/

	elapsed := time.Since(start)
	log.Printf("Delta update took %s", &elapsed)

	return nil

}

func (ob *OrderBook) UpdateBook(bids [][]string, asks [][]string) error {
	ob.HandleLevels(bids, ob.Bids, false)
	ob.HandleLevels(asks, ob.Asks, false)

	return nil
}

func (ob *OrderBook) HandleLevels(levels [][]string, book map[float64]float64, isDelete bool) {

	for _, level := range levels {
		price, err := strconv.ParseFloat(level[0], 64)
		if err != nil {
			log.Println("error parsing price", err)
		}
		qty, err := strconv.ParseFloat(level[1], 64)
		if err != nil {
			log.Println("error parsing qty", err)
		}

		if isDelete || qty == 0 {
			delete(book, price)

		} else {
			book[price] = qty
		}
	}
}

func CalculateMidPrice(bid, ask float64) (float64, error) {

	if bid <= 0 || ask <= 0 {
		return 0, errors.New("invalid input")
	}
	return (bid + ask) / 2, nil

}

func (ob *OrderBook) LogMidPrice() error {

	var bestBid, bestAsk float64

	for price := range ob.Bids {
		if bestBid == 0 || price < bestBid {
			bestBid = price
		}
	}

	for price := range ob.Asks {
		if bestAsk == 0 || price < bestAsk {
			bestAsk = price
		}
	}

	midprice, err := CalculateMidPrice(bestBid, bestAsk)
	if err != nil {
		return err
	}

	log.Printf("Midprice %f", midprice)

	return nil

}
