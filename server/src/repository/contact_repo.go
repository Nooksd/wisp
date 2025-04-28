package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ContactRepo struct{ col *mongo.Collection }

func NewContactRepo(db *mongo.Database) *ContactRepo {
	return &ContactRepo{col: db.Collection("contacts")}
}

func (r *ContactRepo) EnsureDoc(ctx context.Context, ownerID string) error {
	now := time.Now()
	_, err := r.col.UpdateOne(ctx,
		bson.M{"ownerId": ownerID},
		bson.M{"$setOnInsert": bson.M{"ownerId": ownerID, "createdAt": now, "updatedAt": now}},
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *ContactRepo) AddContact(ctx context.Context, ownerID, contactID string) error {
	if err := r.EnsureDoc(ctx, ownerID); err != nil {
		return err
	}
	_, err := r.col.UpdateOne(ctx,
		bson.M{"ownerId": ownerID},
		bson.M{"$addToSet": bson.M{"contactIds": contactID},
			"$set": bson.M{"updatedAt": time.Now()}},
	)
	return err
}

func (r *ContactRepo) RemoveContact(ctx context.Context, ownerID, contactID string) error {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"ownerId": ownerID},
		bson.M{"$pull": bson.M{"contactIds": contactID},
			"$set": bson.M{"updatedAt": time.Now()}},
	)
	return err
}
