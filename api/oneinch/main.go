package oneinch

import _ "embed"

const (
	apiURL      = "https://api.1inch.dev"
	apiRouterV3 = "0x1111111254eeb25477b68fb85ed929f73a960582"
	apiRouterV4 = "0x111111125421cA6dc452d289314280a0f8842A65"
)

//go:embed 1inch.api.key
var apiKey string
