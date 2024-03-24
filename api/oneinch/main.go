package oneinch

import _ "embed"

const (
	apiURL    = "https://api.1inch.dev"
	apiRouter = "0x1111111254eeb25477b68fb85ed929f73a960582"
)

//go:embed 1inch.api.key
var apiKey string
