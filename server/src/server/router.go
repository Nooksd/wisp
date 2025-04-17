package server

import (
	"time"
	"wisp/config"
	"wisp/src/handler"
	"wisp/src/middleware"
	"wisp/src/repository"
	"wisp/src/routes"
	"wisp/src/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewRouter(cfg *config.Config, logger zerolog.Logger, mongoClient *mongo.Client) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	corsCfg := cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsCfg))

	userRepo := repository.NewUserRepo(mongoClient.Database(cfg.Mongo.DBName))
	userSvc := service.NewUserService(userRepo)
	authSvc := service.NewAuthService(mongoClient.Database(cfg.Mongo.DBName), cfg)

	contactRepo := repository.NewContactRepo(mongoClient.Database(cfg.Mongo.DBName))
	frRepo := repository.NewFriendRequestRepo(mongoClient.Database(cfg.Mongo.DBName))
	contactSvc := service.NewContactService(userRepo, contactRepo, frRepo, mongoClient.Database(cfg.Mongo.DBName))

	authHandler := handler.NewAuthHandler(authSvc, userSvc)
	userHandler := handler.NewUserHandler(userSvc)
	contactHandler := handler.NewContactHandler(contactSvc)

	public := r.Group("/")
	secure := r.Group("/")
	secure.Use(middleware.JWTAuth(authSvc))

	routes.UserRoutes(secure, public, userHandler)
	routes.AuthRoutes(secure, public, authHandler)
	routes.ContactRoutes(secure, contactHandler)

	return r
}
