package main

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	database "alert-service/database/sqlc"

	"nhooyr.io/websocket"
)

type SubscribeResponse struct {
	Result interface{} `json:"result"`
	Id     int         `json:"id"`
}

type StreamData struct {
	Price string `json:"p"`
}

type StreamResponse struct {
	Stream string     `json:"stream"`
	Data   StreamData `json:"data"`
}

type cryptoWatcher struct {
	market     *SafeMap
	currencies []currency
	ws         *websocket.Conn
	errch      chan error

	// throttling my market readers for demo purposes
	ticker *time.Ticker

	cache    Cacher
	db       database.Querier
	producer Producer
}

func NewCryptoWatcher(ctx context.Context, currencies []currency, cache Cacher, db database.Querier, producer Producer) (*cryptoWatcher, error) {
	c, _, err := websocket.Dial(ctx, "wss://stream.binance.com/stream", nil)
	if err != nil {
		return nil, err
	}

	safemap := NewSafeMap()

	// map init
	for _, curr := range currencies {
		safemap.Set(curr, "0")
	}

	// prepring subscribe request payload
	subscribePayload := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": currencies,
		"id":     1,
	}
	payloadBytes, err := json.Marshal(subscribePayload)
	if err != nil {
		return nil, err
	}

	// sending subscribe request
	err = c.Write(ctx, websocket.MessageText, payloadBytes)
	if err != nil {
		return nil, err
	}

	// reading subscribe response and checking if subscription was successful
	_, p, err := c.Read(ctx)
	if err != nil {
		return nil, err
	}
	var pubResponse SubscribeResponse
	err = json.Unmarshal(p, &pubResponse)
	if err != nil {
		return nil, err
	}
	log.Println(pubResponse)
	if pubResponse.Result != nil {
		return nil, ErrSubscriptionFailed
	}

	return &cryptoWatcher{
		market:   safemap,
		currencies: currencies,
		ws:       c,
		errch:    make(chan error),
		ticker:   time.NewTicker(100 * time.Millisecond),
		cache:    cache,
		db:       db,
		producer: producer,
	}, nil
}

func (c *cryptoWatcher) Close() error {
	return c.ws.Close(websocket.StatusNormalClosure, "")
}

func (c *cryptoWatcher) Run(ctx context.Context) error {
	// todo unmarshall and fill the market
	go c.fillMarket(ctx)

	// start comparing with target price of users
	for _, curr := range c.currencies {
		go c.startComparing(ctx, curr)
	}

	// handles errors, can be a potential centalized thingy
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case err := <-c.errch:
			logger.Error().Str("err", err.Error())
		}
	}
}

func (c *cryptoWatcher) fillMarket(ctx context.Context) {
	for {
		_, p, err := c.ws.Read(ctx)
		if err != nil {
			c.errch <- err
		}

		// unmarshall and fill the market
		var streamResponse StreamResponse
		err = json.Unmarshal(p, &streamResponse)
		if err != nil {
			c.errch <- err
		}

		c.market.Set(currency(streamResponse.Stream), streamResponse.Data.Price)
		// logger.Info().
		// 	Str("currency", streamResponse.Stream).
		// 	Str("price", streamResponse.Data.Price).
		// 	Send()
	}
}

func (c *cryptoWatcher) startComparing(ctx context.Context, curr currency) {
	// reaading market price after tick time
	for range c.ticker.C {
		price, ok := c.market.Get(curr)
		if !ok {
			logger.Error().
				Str("msg", "unknown currency").
				Send()
			break
		} 
		
		switch price {
		// skips when in memory market is not filled yet
		case "0":
			continue

		default:
			// first get all targets from gt from 0 to current price
			targets, err := c.cache.GetTargets(ctx, curr, true, price)
			if err != nil {
				c.errch <- err
			}

			for _, ID := range targets {
				logger.Info().
					Str("currency", string(curr)).
					Str("price", price).
					Str("alertID", ID).
					Send()

				id, err := strconv.ParseInt(ID, 10, 64)
				if err != nil {
					c.errch <- err
					continue
				}
				params := database.UpdateAlertStatusParams{
					ID:     id,
					Status: string(Triggered),
				}
				err = c.db.UpdateAlertStatus(ctx, params)
				if err != nil {
					c.errch <- err
				}

				// send to kafka
				err = c.producer.Send(ID, price)
				if err != nil {
					c.errch <- err
				}
			}
		}
	}
}
