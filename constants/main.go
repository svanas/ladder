package constants

import "time"

const (
	FLAG_API_KEY     = "api-key"
	FLAG_API_SECRET  = "api-secret"
	FLAG_ASSET       = "asset"
	FLAG_QUOTE       = "quote"
	START_AT_PRICE   = "start-at-price"
	STOP_AT_PRICE    = "stop-at-price"
	START_WITH_SIZE  = "start-with-size"
	FLAG_MULT        = "mult"
	FLAG_SIZE        = "size"
	FLAG_EXCHANGE    = "exchange"
	FLAG_DRY_RUN     = "dry-run"
	FLAG_SIDE        = "side"
	FLAG_CHAIN_ID    = "chain-id"
	FLAG_PRIVATE_KEY = "private-key"
	FLAG_CANCEL      = "cancel"
	FLAG_DAYS        = "days"
)

const (
	ONE_YEAR    = 365 * 24 * time.Hour
	TWO_YEARS   = ONE_YEAR * 2
	THREE_YEARS = ONE_YEAR * 3
)
