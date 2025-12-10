package api

import (
	"context"
	"log"

	mongo "github.com/mwdev22/database/mongo"
	config "github.com/mwdev22/gocfg"
	"github.com/mwdev22/grpclib/grpcserver"
)

type Api struct {
	cfg     *config.Config
	server  *grpcserver.Server
	mongoDB *mongo.MongoDB
}

func New(cfg *config.Config, server *grpcserver.Server, opts ...func(*Api)) *Api {
	api := &Api{
		cfg:    cfg,
		server: server,
	}
	for _, opt := range opts {
		opt(api)
	}
	return api
}

func WithMongoDB(m *mongo.MongoDB) func(*Api) {
	return func(a *Api) {
		a.mongoDB = m
	}
}

func (a *Api) Start(ctx context.Context) error {

	// a.server.RegisterService()

	addr, err := a.server.Start(ctx)
	if err != nil {
		log.Fatalf("couldn't run grpc server: %s", err)
	}

	log.Printf("listening on %s", addr)

	return nil
}
