package gateway

import config "github.com/mwdev22/gocfg"

type AppConfig struct {
	Base     *config.Config
	BookingAddr string
	PaymentAddr string
}
