package server

import (
	"time"
	"wisp/config"
	"wisp/src/handler"
	"wisp/src/middleware"
	"wisp/src/repository"
	"wisp/src/routes"
	"wisp/src/service"
	"wisp/src/ws"

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

	db := mongoClient.Database(cfg.Mongo.DBName)

	// Repositórios
	userRepo := repository.NewUserRepo(db)
	contactRepo := repository.NewContactRepo(db)
	frRepo := repository.NewFriendRequestRepo(db)
	msgRepo := repository.NewMessageRepo(db)

	// Serviços
	userSvc := service.NewUserService(userRepo)
	authSvc := service.NewAuthService(db, cfg)
	contactSvc := service.NewContactService(userRepo, contactRepo, frRepo, db)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc, userSvc)
	userHandler := handler.NewUserHandler(userSvc)
	contactHandler := handler.NewContactHandler(contactSvc)

	// WebSocket Hub
	hub := ws.NewHub(msgRepo)
	go hub.Run() // Inicia o hub em uma goroutine separada

	// WebSocket Handler
	wsHandler := handler.NewWSHandler(hub)

	public := r.Group("/")
	secure := r.Group("/")
	secure.Use(middleware.JWTAuth(authSvc))

	// Configuração das rotas
	routes.UserRoutes(secure, public, userHandler)
	routes.AuthRoutes(secure, public, authHandler)
	routes.ContactRoutes(secure, contactHandler)
	routes.WSRoutes(secure, wsHandler)

	return r
}
