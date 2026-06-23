package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	QueueName           = "luxus-connect-events"
	ExchangeName        = "luxus-connect-events"
	ImportRoutingPrefix = "providers.invoice_import.requested"
)

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	log  *slog.Logger
}

type InvoiceImportRequestedEvent struct {
	AggregateID        string  `json:"aggregate_id"`
	EventType          string  `json:"event_type"`
	StorageBucket      string  `json:"storage_bucket"`
	StorageObjectKey   string  `json:"storage_object_key"`
	OriginalFileName   *string `json:"original_file_name"`
	RequestedByUserID  string  `json:"requested_by_user_id"`
}

func NewPublisher(rabbitURL string, log *slog.Logger) (*Publisher, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}
	if err := ch.ExchangeDeclare(ExchangeName, "topic", true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	return &Publisher{conn: conn, ch: ch, log: log}, nil
}

func (p *Publisher) Close() {
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
}

func (p *Publisher) PublishInvoiceImportRequested(ctx context.Context, importRequestID, bucket, key string, originalFileName *string, userID string) error {
	if p == nil || p.ch == nil {
		return fmt.Errorf("rabbitmq publisher not available")
	}
	evt := InvoiceImportRequestedEvent{
		AggregateID:       importRequestID,
		EventType:         "InvoiceImportRequestedEvent",
		StorageBucket:     bucket,
		StorageObjectKey:  key,
		OriginalFileName:  originalFileName,
		RequestedByUserID: userID,
	}
	body, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	routingKey := fmt.Sprintf("%s.%s", ImportRoutingPrefix, importRequestID)
	return p.ch.PublishWithContext(ctx, ExchangeName, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

type ImportProcessor interface {
	ProcessImport(ctx context.Context, importRequestID string) error
}

type Consumer struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	processor ImportProcessor
	log       *slog.Logger
}

func NewConsumer(rabbitURL string, processor ImportProcessor, log *slog.Logger) (*Consumer, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	if err := ch.ExchangeDeclare(ExchangeName, "topic", true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	if _, err := ch.QueueDeclare(QueueName, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	if err := ch.QueueBind(QueueName, ImportRoutingPrefix+".#", ExchangeName, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	return &Consumer{conn: conn, ch: ch, processor: processor, log: log}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.ch.Consume(QueueName, "luxus-connect-api", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}
				c.handleDelivery(ctx, d)
			}
		}
	}()
	return nil
}

func (c *Consumer) handleDelivery(ctx context.Context, d amqp.Delivery) {
	var evt InvoiceImportRequestedEvent
	if err := json.Unmarshal(d.Body, &evt); err != nil {
		c.log.Error("invalid event payload", "error", err)
		_ = d.Nack(false, false)
		return
	}
	if err := c.processor.ProcessImport(ctx, evt.AggregateID); err != nil {
		c.log.Error("import processing failed", "import_request_id", evt.AggregateID, "error", err)
	}
	_ = d.Ack(false)
}

func (c *Consumer) Close() {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}
