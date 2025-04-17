package service

import (
	"context"
	"errors"
	"time"

	"wisp/config"
	"wisp/src/model"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	usersCol    *mongo.Collection
	sessionsCol *mongo.Collection
	jwtSecret   string
	ttl         time.Duration
}

type Claims struct {
	UserID  string `json:"userId"`
	IsAdmin bool   `json:"isAdmin"`
	jwt.RegisteredClaims
}

func NewAuthService(db *mongo.Database, cfg *config.Config) *AuthService {
	return &AuthService{
		usersCol:    db.Collection("users"),
		sessionsCol: db.Collection("sessions"),
		jwtSecret:   cfg.App.Env + "_" + cfg.Jwt.Secret,
		ttl:         7 * 24 * time.Hour,
	}
}

func (a *AuthService) Login(ctx context.Context, email, pass, deviceID string) (string, error) {
	var u model.User
	if err := a.usersCol.FindOne(ctx, bson.M{"email": email}).Decode(&u); err != nil {
		return "", errors.New("credenciais inválidas")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pass)); err != nil {
		return "", errors.New("credenciais inválidas")
	}

	a.sessionsCol.DeleteMany(ctx, bson.M{"userId": u.ID.Hex()})

	now := time.Now()
	claims := Claims{
		UserID:  u.UserID,
		IsAdmin: u.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        primitive.NewObjectID().Hex(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		return "", err
	}

	_, err = a.sessionsCol.InsertOne(ctx, bson.M{
		"sid":      claims.ID,
		"userId":   u.ID.Hex(),
		"deviceId": deviceID,
		"expires":  claims.ExpiresAt.Time,
	})
	return signed, err
}

func (a *AuthService) Logout(ctx context.Context, sid string) error {
	_, err := a.sessionsCol.DeleteOne(ctx, bson.M{"sid": sid})
	return err
}

func (a *AuthService) ValidateToken(tokenStr string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return []byte(a.jwtSecret), nil
	})

	if err != nil || !tok.Valid {
		return nil, errors.New("token inválido")
	}

	claims := tok.Claims.(*Claims)

	count, err := a.sessionsCol.CountDocuments(context.Background(), bson.M{"sid": claims.ID})
	if err != nil || count == 0 {
		return nil, errors.New("sessão expirada")
	}
	return claims, nil
}
