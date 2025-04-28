// internal/service/contact_service.go
package service

import (
	"context"
	"errors"
	"wisp/src/model"
	"wisp/src/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ContactService struct {
	userRepo    *repository.UserRepo
	contactRepo *repository.ContactRepo
	frRepo      *repository.FriendRequestRepo
	usersCol    *mongo.Collection
	contactCol  *mongo.Collection
	frCol       *mongo.Collection
}

func NewContactService(
	ur *repository.UserRepo,
	cr *repository.ContactRepo,
	fr *repository.FriendRequestRepo,
	db *mongo.Database,
) *ContactService {
	return &ContactService{
		userRepo:    ur,
		usersCol:    db.Collection("users"),
		contactRepo: cr,
		contactCol:  db.Collection("contacts"),
		frRepo:      fr,
		frCol:       db.Collection("friend_requests"),
	}
}

func (s *ContactService) GetContacts(ctx context.Context, userUID string) ([]map[string]any, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"ownerId": userUID}}},
		{{Key: "$unwind", Value: "$contactIds"}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "contactIds",
			"foreignField": "userId",
			"as":           "user",
		}}},
		{{Key: "$unwind", Value: "$user"}},
		{{Key: "$replaceRoot", Value: bson.M{"newRoot": "$user"}}},
		{{Key: "$project", Value: bson.M{
			"_id":    0,
			"userId": 1,
			"name":   1,
			"email":  1,
		}}},
	}

	cur, err := s.contactCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []map[string]any
	for cur.Next(ctx) {
		var doc map[string]any
		cur.Decode(&doc)
		results = append(results, doc)
	}

	return results, nil
}

func (s *ContactService) GetIncomingRequests(ctx context.Context, userUID string) ([]model.FriendRequest, error) {
	return s.frRepo.GetIncoming(ctx, userUID)
}

func (s *ContactService) GetSentRequests(ctx context.Context, userUID string) ([]model.FriendRequest, error) {
	return s.frRepo.GetSent(ctx, userUID)
}

func (s *ContactService) SendFriendRequest(ctx context.Context, fromUID, toUID string) (primitive.ObjectID, error) {
	to, err := s.userRepo.FindByUserID(ctx, toUID)
	if err != nil {
		return primitive.NilObjectID, errors.New("usuário alvo não existe")
	}

	return s.frRepo.Create(ctx, fromUID, to.UserID)
}

func (s *ContactService) CancelFriendRequest(ctx context.Context, reqID string, userUID string) error {
	id, _ := primitive.ObjectIDFromHex(reqID)
	fr, err := s.frRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if fr.FromUserID != userUID {
		return errors.New("não autorizado")
	}

	return s.frRepo.DeleteByID(ctx, id)
}

func (s *ContactService) AcceptFriendRequest(ctx context.Context, reqID, userUID string) error {
	id, _ := primitive.ObjectIDFromHex(reqID)
	fr, err := s.frRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if fr.ToUserID != userUID {
		return errors.New("não autorizado")
	}

	if err := s.contactRepo.AddContact(ctx, fr.FromUserID, fr.ToUserID); err != nil {
		return err
	}
	if err := s.contactRepo.AddContact(ctx, fr.ToUserID, fr.FromUserID); err != nil {
		return err
	}
	return s.frRepo.DeleteByID(ctx, id)
}

func (s *ContactService) RejectFriendRequest(ctx context.Context, reqID, userUID string) error {
	id, _ := primitive.ObjectIDFromHex(reqID)
	fr, err := s.frRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if fr.ToUserID != userUID {
		return errors.New("não autorizado")
	}
	return s.frRepo.DeleteByID(ctx, id)
}

func (s *ContactService) RemoveContact(ctx context.Context, userUID, targetUID string) error {
	me, err := s.userRepo.FindByUserID(ctx, userUID)
	if err != nil {
		return err
	}
	tgt, err := s.userRepo.FindByUserID(ctx, targetUID)
	if err != nil {
		return err
	}

	if err := s.contactRepo.RemoveContact(ctx, me.UserID, tgt.UserID); err != nil {
		return err
	}
	return s.contactRepo.RemoveContact(ctx, tgt.UserID, me.UserID)
}
