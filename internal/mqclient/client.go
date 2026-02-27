// Package mqclient wraps amqp091 to subscribe to the GeoKrety RabbitMQ
// fanout exchange and dispatch incoming move events to the scoring engine.
package mqclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

	"github.com/geokrety/geokrety-points-system/internal/config"
)

// incomingMsg is the JSON structure of messages on the geokrety exchange.
// Example: {"id": 12345, "op": "INSERT", "kind": "gk_moves"}
type incomingMsg struct {
	ID   int64  `json:"id"`
	Op   string `json:"op"`
	Kind string `json:"kind"`
}

// Handler is called for each incoming move ID.
type Handler func(ctx context.Context, moveID int64) error

// Client subscribes to the GeoKrety RabbitMQ fanout exchange.
type Client struct {
	cfg       config.Config
	handler   Handler
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

// New creates a new AMQP client. Call Start() to begin receiving messages.
func New(cfg config.Config, handler Handler) *Client {
	return &Client{cfg: cfg, handler: handler}
}

// Start connects to RabbitMQ and begins consuming messages.
// It blocks until ctx is cancelled, reconnecting on transient failures.
func (c *Client) Start(ctx context.Context) error {
	for {
		err := c.connect()
		if err != nil {
			log.Error().Err(err).Msg("AMQP connect failed; retrying in 5s")
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(5 * time.Second):
				continue
			}
		}

		log.Info().Str("exchange", c.cfg.AMQP.Exchange).Msg("AMQP connected")

		if err := c.consume(ctx); err != nil {
			log.Error().Err(err).Msg("AMQP consume loop exited; reconnecting")
		}

		c.closeConnection()

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(2 * time.Second):
		}
	}
}

// connect dials RabbitMQ, declares the exchange, and sets up an exclusive queue.
func (c *Client) connect() error {
	conn, err := amqp.Dial(c.cfg.AMQP.URL())
	if err != nil {
		return fmt.Errorf("amqp dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}

	// Declare the exchange (passive — we don't own it).
	if err := ch.ExchangeDeclarePassive(
		c.cfg.AMQP.Exchange,
		"fanout",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("exchange declare: %w", err)
	}

	// Create an exclusive anonymous queue bound to the fanout exchange.
	q, err := ch.QueueDeclare(
		"",    // name: "" = broker-generated
		false, // durable
		true,  // auto-delete
		true,  // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("queue declare: %w", err)
	}

	if err := ch.QueueBind(q.Name, "", c.cfg.AMQP.Exchange, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("queue bind: %w", err)
	}

	c.conn = conn
	c.channel = ch
	c.queueName = q.Name
	return nil
}

// consume reads messages from the queue until ctx is cancelled or the channel closes.
func (c *Client) consume(ctx context.Context) error {
	deliveries, err := c.channel.Consume(
		c.queueName, // queue (auto-assigned during connect)
		"",          // consumer tag: broker-assigned
		false,       // auto-ack: manual ack so we don't lose messages on crash
		true,        // exclusive
		false,       // no-local
		false,       // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("delivery channel closed")
			}
			c.handle(ctx, d)
		}
	}
}

// handle processes one delivery.
func (c *Client) handle(ctx context.Context, d amqp.Delivery) {
	var msg incomingMsg
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		log.Warn().Bytes("body", d.Body).Err(err).Msg("AMQP: invalid JSON, discarding")
		_ = d.Ack(false)
		return
	}

	// Only process INSERT events on gk_moves.
	if msg.Op != "INSERT" || msg.Kind != "gk_moves" {
		_ = d.Ack(false)
		return
	}

	if err := c.handler(ctx, msg.ID); err != nil {
		log.Error().
			Int64("move_id", msg.ID).
			Err(err).
			Msg("AMQP: handler error; nacking (requeue=false)")
		_ = d.Nack(false, false)
		return
	}

	_ = d.Ack(false)
}

// closeConnection closes the AMQP channel and connection if open.
func (c *Client) closeConnection() {
	if c.channel != nil {
		_ = c.channel.Close()
		c.channel = nil
	}
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}
