package main

import (
	"encoding/json"
	"flag"
	"fmt"
	_ "io/ioutil"
	"net/http"
	"time"
)

const (
	HEARTBEAT_URL       = "https://api.stockfighter.io/ob/api/heartbeat"
	VENUE_HEARTBEAT_URL = "https://api.stockfighter.io/ob/api/venues/%s/heartbeat"
	VENUE_STOCKS_URL    = "https://api.stockfighter.io/ob/api/venues/%s/stocks"
	VENUE_ORDERS_URL    = "https://api.stockfighter.io/ob/api/venues/%s/stocks/%s"
	VENUE_POST_ORDER    = "https://api.stockfighter.io/ob/api/venues/%s/stocks/%s/orders"
)

type Direction string

const (
	BUY  Direction = "buy"
	SELL Direction = "sell"
)

type OrderType string

const (
	Limit  OrderType = "limit"
	Market OrderType = "market"
	FOK    OrderType = "fill-or-kill"
	IOC    OrderType = "immediate-or-cancel"
)

type Stockfighter interface {
	Heartbeat() bool
	VenueHeartbeat(venue string) bool
}

type HasOk struct {
	Ok bool `json:"ok"`
}

type StatusResponse struct {
	HasOk
	Error string `json:"error"`
}

type StocksResponse struct {
	HasOk
	Symbols []Symbol `json:"symbols"`
}

type Symbol struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

type Order struct {
	Price    uint64 `json:"price"`
	Quantity uint64 `json:"qty"`
}

type OrderResponse struct {
	HasOk
	Symbol
	Bids      []OrderbookOrder `json:"bids"`
	Asks      []OrderbookOrder `json:"asks"`
	Timestamp time.Time        `json:"ts"`
}

type OrderbookOrder struct {
	Order
	IsBuy bool `json:"isBuy"`
}

type OrderRequest struct {
	Account   string `json:"account"`
	OrderType string `json:"orderType"`
	Direction string `json:"direction"`
	Qty       int    `json:"qty"`
	Price     int    `json:"price"`
}

type Quote struct {
	HasOk
	Symbol

	Bid      int `json:"bid"`
	BidSize  int `json:"bidSize"`
	BidDepth int `json:"bidDepth"`

	AskSize  int `json:"askSize"`
	AskDepth int `json:"askDepth"`

	Last     int `json:"last"`
	LastSize int `json:"lastSize"`

	LastTrade time.Time `json:"lastTrade"`
	QuoteTime time.Time `json:"quoteTime"`
}

type Client struct {
	apiKey string
}

func (c Client) Heartbeat() bool {
	fmt.Println("Welcome!")

	resp, err := http.Get(HEARTBEAT_URL)

	if err != nil {
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	s := new(StatusResponse)
	err = json.NewDecoder(resp.Body).Decode(&s)

	if s.Ok != true {
		return false
	}

	return true
}

func (c Client) VenueHeartbeat(venue string) bool {
	resp, err := http.Get(fmt.Sprintf(VENUE_HEARTBEAT_URL, venue))

	if err != nil {
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	s := new(StatusResponse)
	err = json.NewDecoder(resp.Body).Decode(&s)

	if s.Ok != true {
		return false
	}

	return true
}

func (c Client) VenueStocks(venue string) []Symbol {
	resp, err := http.Get(fmt.Sprintf(VENUE_STOCKS_URL, venue))

	if err != nil {
		return nil
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil
	}

	stocks := new(StocksResponse)
	err = json.NewDecoder(resp.Body).Decode(stocks)

	if err != nil {
		return nil
	}

	return stocks.Symbols
}

func (c Client) VenueOrders(venue string, stock string) *OrderResponse {
	resp, err := http.Get(fmt.Sprintf(VENUE_ORDERS_URL, venue, stock))

	if err != nil {
		return nil
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil
	}

	if resp.StatusCode != 200 {
		return nil
	}

	orders := new(OrderResponse)
	err = json.NewDecoder(resp.Body).Decode(orders)

	return orders
}

func (c Client) PlaceOrder(venue string, stock string, account string, price int, qty int, dir Direction, otype OrderType) {

	or := OrderRequest{account, string(otype), string(dir), qty, price}
	or = or
	// json.NewEncoder(or)
	// payload, err := json.Marshal(or)

	// if err != nil {
	// 	return
	// }

	// resp, err := http.Post(fmt.Sprintf(VENUE_POST_ORDER, venue, stock), payload)
}

func main() {
	var key = flag.String("key", "", "API Key")
	flag.Parse()
	c := Client{*key}

	// ok := c.Heartbeat()
	// ok := c.VenueHeartbeat("TESTEX")
	// stocks := c.VenueStocks("TESTEX")
	orders := c.VenueOrders("TESTEX", "FOOBAR")
	fmt.Println(orders.Timestamp.Date())
}
