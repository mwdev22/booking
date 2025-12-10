package main

import (
	"context"
	"log"

	"github.com/mwdev22/booking/api"
	mongo "github.com/mwdev22/database/mongo"
	config "github.com/mwdev22/gocfg"
	"github.com/mwdev22/grpclib/grpcserver"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := config.New(
		config.WithDatabaseConfig(
			&config.DatabaseConfig{
				URI: config.GetEnv("DB_URI", ""),
			},
		),
	)

	db := mongo.New(cfg.Database.URI)

	server := grpcserver.New(
		cfg.Addr,
		grpcserver.WithCreds(insecure.NewCredentials()),
	)

	app := api.New(cfg, server,
		api.WithMongoDB(
			db,
		),
	)

	ctx := context.Background()

	if err := app.Start(ctx); err != nil {
		log.Fatalf("error while running app: %v", err)
	}
}
