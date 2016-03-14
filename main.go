package main

import (
	"encoding/json"
	"errors"
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
	VENUE_ORDERBOOK_URL = "https://api.stockfighter.io/ob/api/venues/%s/stocks/%s"
	VENUE_QUOTE_URL     = "https://api.stockfighter.io/ob/api/venues/%s/stocks/%s/quote"

	VENUE_ORDERS_URL = "https://api.stockfighter.io/ob/api/venues/%s/stocks/%s/orders"
	VENUE_ORDER_URL  = "https://api.stockfighter.io/ob/api/venues/%s/stocks/%s/orders/%d"

	ACCT_EVERY_ORDERS_URL = "https://api.stockfighter.io/ob/api/venues/%s/accounts/%s/orders"
	ACCT_STOCK_ORDERS_URL = "https://api.stockfighter.io/ob/api/venues/%s/accounts/%s/stocks/%s/orders"
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
	Ping() bool

	PingVenue(venue string) bool

	FetchStocks(venue string) []SymbolInfo
	FetchOrderbook(venue string, stock string) *OrderbookResponse
	FetchQuote(venue string, stock string) *QuoteResponse

	FetchOrder(venue string, stock string, id int) *OrderResponse
	PlaceOrder(venue string, stock string, order OrderRequest) *OrderResponse
	CancelOrder(venue string, stock string, id int) bool

	FetchAcctOrders(venue string, account string) []OrderResponse
	FetchAcctStockOrders(venue string, stock string, account string) []OrderResponse

	HandleNewQuote()
	HandleNewOrder()
}

type HasOk struct {
	Ok bool `json:"ok"`
}

type StatusResponse struct {
	HasOk
	Error string `json:"error"`
}

type VenueStatusResponse struct {
	HasOk
	Venue string `json:"venue"`
}

type SymbolInfo struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

type VenueStocksResponse struct {
	HasOk
	Symbols []SymbolInfo `json:"symbols"`
}

type Order struct {
	Price    uint64 `json:"price"`
	Quantity uint64 `json:"qty"`
}

type OrderbookEntry struct {
	Order
	IsBuy bool `json:"isBuy"`
}

type OrderbookResponse struct {
	VenueStatusResponse
	Symbol    string           `json:"symbol"`
	Bids      []OrderbookEntry `json:"bids"`
	Asks      []OrderbookEntry `json:"asks"`
	Timestamp time.Time        `json:"ts"`
}

type OrderRequest struct {
	Account   string `json:"account"`
	OrderType string `json:"orderType"`
	Direction string `json:"direction"`
	Qty       int    `json:"qty"`
	Price     int    `json:"price"`
}

type Fill struct {
	Qty       int       `json:"qty"`
	Price     int       `json:"price"`
	Timestamp time.Time `json:"ts"`
}

type OrderResponse struct {
	VenueStatusResponse
	OrderRequest
	OriginalQty int `json:"originalQty"`

	Id        int       `json:"id"`
	Timestamp time.Time `json:"ts"`

	Fills       []Fill `json:"fills"`
	TotalFilled int    `json:"totalFilled"`

	Open bool `json:"open"`
}

type OrdersStatusResponse struct {
	VenueStatusResponse
	Orders []OrderResponse `json:"orders"`
}

type QuoteResponse struct {
	VenueStatusResponse
	Symbol string `json:"symbol"`

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

func PerformRequest(url string, method string, body *string, js interface{}) error {
	var resp *http.Response
	var err error

	switch {
	case method == http.MethodGet:
		resp, err = http.Get(url)
	case method == http.MethodPost:
		resp, err = http.Post(url, *body, nil)
	case method == http.MethodDelete:
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return err
		}

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
	}

	if err != nil {
		return errors.New(fmt.Sprintf("Request failed: %d", resp.StatusCode))
	}

	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&js); err != nil {
		return err
	}

	return nil
}

func (c Client) Ping() bool {
	s := new(StatusResponse)
	err := PerformRequest(HEARTBEAT_URL, http.MethodGet, nil, &s)

	if err != nil {
		return false
	}

	if s.Ok != true {
		return false
	}

	return true
}

func (c Client) PingVenue(venue string) bool {
	s := new(VenueStatusResponse)
	err := PerformRequest(fmt.Sprintf(VENUE_HEARTBEAT_URL, venue), http.MethodGet, nil, &s)

	if err != nil {
		return false
	}

	if s.Ok != true {
		return false
	}

	return true
}

func (c Client) FetchStocks(venue string) []SymbolInfo {
	s := new(VenueStocksResponse)
	err := PerformRequest(fmt.Sprintf(VENUE_STOCKS_URL, venue), http.MethodGet, nil, &s)

	if err != nil {
		return nil
	}

	if s.Ok != true {
		return nil
	}

	return s.Symbols
}

func (c Client) FetchOrder(venue string, stock string, id int) *OrderResponse {
	s := new(OrderResponse)
	err := PerformRequest(fmt.Sprintf(VENUE_ORDER_URL, venue, stock, id), http.MethodGet, nil, &s)

	if err != nil {
		return nil
	}

	if s.Ok != true {
		return nil
	}

	if s.Venue != venue {
		return nil
	}

	return s
}

func (c Client) FetchOrderbook(venue string, stock string) *OrderbookResponse {
	s := new(OrderbookResponse)
	err := PerformRequest(fmt.Sprintf(VENUE_ORDERBOOK_URL, venue, stock), http.MethodGet, nil, &s)

	if err != nil {
		return nil
	}

	if s.Ok != true {
		return nil
	}

	if s.Venue != venue {
		return nil
	}

	if s.Symbol != stock {
		return nil
	}

	return s
}

func (c Client) FetchAcctOrders(venue string, account string) []OrderResponse {
	s := new(OrdersStatusResponse)
	err := PerformRequest(fmt.Sprintf(ACCT_EVERY_ORDERS_URL, account, venue), http.MethodGet, nil, &s)

	if err != nil {
		return nil
	}

	if s.Ok != true {
		return nil
	}

	if s.Venue != venue {
		return nil
	}

	return s.Orders
}

func (c Client) FetchAcctStockOrders(venue string, stock string, account string) []OrderResponse {
	s := new(OrdersStatusResponse)
	err := PerformRequest(fmt.Sprintf(ACCT_STOCK_ORDERS_URL, account, venue, stock), http.MethodGet, nil, &s)

	if err != nil {
		return nil
	}

	if s.Ok != true {
		return nil
	}

	if s.Venue != venue {
		return nil
	}

	return s.Orders
}

func (c Client) PlaceOrder(venue string, stock string, order OrderRequest) *OrderResponse {
	payload, err := json.Marshal(order)

	if err != nil {
		return nil
	}

	s := new(OrderResponse)
	js := string(payload)
	err = PerformRequest(fmt.Sprintf(VENUE_ORDERS_URL, venue, stock), http.MethodPost, &js, &s)

	return nil
}

func (c Client) FetchQuote(venue string, stock string) *QuoteResponse {
	s := new(QuoteResponse)
	err := PerformRequest(fmt.Sprintf(VENUE_QUOTE_URL, venue, stock), http.MethodGet, nil, &s)

	if err != nil {
		return nil
	}

	if s.Ok != true {
		return nil
	}

	return s
}

func (c Client) CancelOrder(venue string, stock string, id int) bool {
	s := new(OrderResponse)
	err := PerformRequest(fmt.Sprintf(VENUE_ORDER_URL, venue, stock, id), http.MethodDelete, nil, &s)

	if err != nil {
		return false
	}

	if s.Ok != true {
		return false
	}

	if s.Open != false {
		return false
	}

	return true
}

func (c Client) HandleNewQuote() {

}

func (c Client) HandleNewOrder() {

}

func Run(client Stockfighter) {
}

func main() {
	var key = flag.String("key", "", "API Key")
	flag.Parse()
	c := Client{*key}

	Run(c)
}
