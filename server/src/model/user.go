package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"  json:"-"`
	UserID       string             `bson:"userId"         json:"userId"        validate:"required,len=7"`
	Name         string             `bson:"name"           json:"name"          validate:"required,min=3"`
	Email        string             `bson:"email"          json:"email"         validate:"required,email"`
	PasswordHash string             `bson:"passwordHash"   json:"-"`
	IsAdmin      bool               `bson:"isAdmin"        json:"isAdmin"`
	CreatedAt    time.Time          `bson:"createdAt"      json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt"      json:"updatedAt"`
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	UserID   string `json:"userId" validate:"required,len=7"`
	Password string `json:"password" validate:"required,min=5"`
}
