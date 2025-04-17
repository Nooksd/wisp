package repository

import (
	"context"
	"time"
	"wisp/src/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FriendRequestRepo struct{ col *mongo.Collection }

func NewFriendRequestRepo(db *mongo.Database) *FriendRequestRepo {
	return &FriendRequestRepo{col: db.Collection("friend_requests")}
}

func (r *FriendRequestRepo) Create(ctx context.Context, from, to string) (primitive.ObjectID, error) {
	fr := model.FriendRequest{
		ID:         primitive.NewObjectID(),
		FromUserID: from,
		ToUserID:   to,
		CreatedAt:  time.Now(),
	}

	res, err := r.col.InsertOne(ctx, fr)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return res.InsertedID.(primitive.ObjectID), nil
}

func (r *FriendRequestRepo) GetIncoming(ctx context.Context, to string) ([]model.FriendRequest, error) {
	cur, err := r.col.Find(ctx, bson.M{"toUserId": to})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []model.FriendRequest
	for cur.Next(ctx) {
		var fr model.FriendRequest
		cur.Decode(&fr)
		out = append(out, fr)
	}
	return out, nil
}

func (r *FriendRequestRepo) GetSent(ctx context.Context, from string) ([]model.FriendRequest, error) {
	cur, err := r.col.Find(ctx, bson.M{"fromUserId": from})
	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)

	var out []model.FriendRequest
	for cur.Next(ctx) {
		var fr model.FriendRequest
		cur.Decode(&fr)
		out = append(out, fr)
	}

	return out, nil
}

func (r *FriendRequestRepo) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *FriendRequestRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.FriendRequest, error) {
	var fr model.FriendRequest
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&fr)
	if err != nil {
		return nil, err
	}

	return &fr, nil
}
