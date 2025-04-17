package repository

import (
	"context"
	"errors"
	"time"

	"wisp/src/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepo struct{ col *mongo.Collection }

func NewUserRepo(db *mongo.Database) *UserRepo {
	return &UserRepo{col: db.Collection("users")}
}

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	now := time.Now()

	u.ID = primitive.NewObjectID()
	u.CreatedAt = now
	u.UpdatedAt = now
	_, err := r.col.InsertOne(ctx, u)
	return err
}

func (r *UserRepo) UpdateByUserID(ctx context.Context, userID string, upd map[string]interface{}) error {
	upd["updatedAt"] = time.Now()
	_, err := r.col.UpdateOne(ctx,
		bson.M{"userId": userID},
		bson.M{"$set": upd},
	)
	return err
}

func (r *UserRepo) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"userId": userID})
	return err
}

func (r *UserRepo) List(ctx context.Context, page, limit int, q, sortField string, desc bool) ([]model.User, int64, error) {
	filter := bson.M{}
	if q != "" {
		regex := bson.M{"$regex": q, "$options": "i"}
		filter["$or"] = bson.A{
			bson.M{"userId": regex},
			bson.M{"name": regex},
			bson.M{"email": regex},
		}
	}
	findOpts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: sortField, Value: func() int {
			if desc {
				return -1
			}
			return 1
		}()}})
	cur, err := r.col.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)
	var users []model.User
	for cur.Next(ctx) {
		var u model.User
		if err := cur.Decode(&u); err == nil {
			users = append(users, u)
		}
	}
	total, _ := r.col.CountDocuments(ctx, filter)
	return users, total, nil
}

func (r *UserRepo) IsAvailable(ctx context.Context, email, userID string) (bool, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{
		"$or": []bson.M{{"email": email}, {"userId": userID}},
	})
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (r *UserRepo) FindByUserID(ctx context.Context, userID string) (*model.User, error) {
	var u model.User
	err := r.col.FindOne(ctx, bson.M{"userId": userID}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("not found")
	}
	return &u, err
}
