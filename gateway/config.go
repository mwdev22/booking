package gateway

import config "github.com/mwdev22/gocfg"

type AppConfig struct {
	BaseCfg     *config.Config
	BookingAddr string
	PaymentAddr string
}
