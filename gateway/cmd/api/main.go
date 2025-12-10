package main

import (
	"log"

	gateway "github.com/mwdev22/booking"
	"github.com/mwdev22/booking/api"
	config "github.com/mwdev22/gocfg"
)

// @title           REST Boilerplate API
// @version         1.0
// @description     API documentation

func main() {
	baseCfg := config.New(
		config.WithDatabaseConfig(
			&config.DatabaseConfig{},
		),
	)

	appCfg := &gateway.AppConfig{
		BaseCfg:     baseCfg,
		BookingAddr: config.GetEnv("BOOKING_ADDR", "localhost:50037"),
		PaymentAddr: config.GetEnv("PAYMENT_ADDR", "localhost:50052"),
	}

	app := api.New(appCfg)

	if err := app.Run(); err != nil {
		log.Fatalf("error while running app: %v", err)
	}
}
