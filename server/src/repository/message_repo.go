package repository

import (
	"context"
	"time"
	"wisp/src/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageRepo struct {
	col *mongo.Collection
}

func NewMessageRepo(db *mongo.Database) *MessageRepo {
	col := db.Collection("pending_messages")
	idx := mongo.IndexModel{
		Keys:    bson.D{{Key: "createdAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(7 * 24 * 3600),
	}
	col.Indexes().CreateOne(context.Background(), idx)
	return &MessageRepo{col: col}
}

func (r *MessageRepo) Insert(ctx context.Context, pm *model.PendingMessage) (primitive.ObjectID, error) {
	pm.CreatedAt = time.Now()
	res, err := r.col.InsertOne(ctx, pm)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

func (r *MessageRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *MessageRepo) GetPendingFor(ctx context.Context, to string) ([]model.PendingMessage, error) {
	cur, err := r.col.Find(ctx, bson.M{"to": to})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var msgs []model.PendingMessage
	for cur.Next(ctx) {
		var pm model.PendingMessage
		if err := cur.Decode(&pm); err == nil {
			msgs = append(msgs, pm)
		}
	}
	return msgs, nil
}
