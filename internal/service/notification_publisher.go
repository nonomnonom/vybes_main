package service

import (
	"encoding/json"
	"vybes/internal/config"
	"vybes/internal/domain"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

const NotificationSubject = "notifications.create"

// NotificationPublisher defines the interface for publishing notification events.
type NotificationPublisher interface {
	Publish(event domain.Notification)
}

type natsNotificationPublisher struct {
	nc *nats.Conn
}

// NewNATSNotificationPublisher creates a new NATS notification publisher.
func NewNATSNotificationPublisher(cfg *config.Config) (NotificationPublisher, error) {
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		return nil, err
	}
	return &natsNotificationPublisher{nc: nc}, nil
}

func (p *natsNotificationPublisher) Publish(event domain.Notification) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal notification event")
		return
	}
	if err := p.nc.Publish(NotificationSubject, data); err != nil {
		log.Error().Err(err).Msg("Failed to publish notification event")
	}
}
