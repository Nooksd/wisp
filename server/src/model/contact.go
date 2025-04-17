package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Contact struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	OwnerID   string             `bson:"ownerId" json:"ownerId" validate:"required"`
	ContactID []string           `bson:"contactId" json:"contactIds"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}
