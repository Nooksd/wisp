package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PendingMessage struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	From      string             `bson:"from"`
	To        string             `bson:"to"`
	Payload   string             `bson:"payload"`
	CreatedAt time.Time          `bson:"createdAt"`
}
