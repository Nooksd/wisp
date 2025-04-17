package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FriendRequest struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	FromUserID string             `bson:"fromUserId" json:"fromUserId" validate:"required,len=7"`
	ToUserID   string             `bson:"toUserId" json:"toUserId" validate:"required,len=7"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
}
