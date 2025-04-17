package main

import (
	"context"
	"fmt"
	"os"
	"wisp/config"
	"wisp/src/db"
	"wisp/src/logger"
	"wisp/src/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "falha ao carregar config: %v\n", err)
		os.Exit(1)
	}

	log := logger.NewLogger(cfg.App.Env)
	log.Info().Msg("starting wisp backend")

	mongoClient, err := db.Connect(cfg.Mongo.URI)
	if err != nil {
		log.Fatal().Err(err).Msg("Não foi possível conectar ao MongoDB")
	}
	defer mongoClient.Disconnect(context.TODO())
	log.Info().Str("db", cfg.Mongo.DBName).Msg("Conectado ao MongoDB")

	r := server.NewRouter(cfg, log, mongoClient)

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Info().Str("addr", addr).Msg("Servidor iniciado")
	if err := r.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("servidor falhou ao iniciar")
	}
}
