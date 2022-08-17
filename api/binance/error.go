package binance

import (
	"context"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/common"
)

func isBinanceError(err error) (*common.APIError, bool) {
	if err == nil {
		return nil, false
	}
	apiError, ok := err.(*common.APIError)
	if ok {
		return apiError, true
	}
	return nil, false
}

// You can ignore this error and continue with the next for-loop iteration
type errorContinue struct{}

// This error is silent by design
func (err *errorContinue) Error() string {
	return ""
}

func handleRecvWindowError(client *binance.Client, err error) error {
	apiError, ok := isBinanceError(err)
	if ok {
		if apiError.Code == -1021 {
			// Timestamp for this request is outside of the recvWindow.
			beforeRequest(client, serverTime)
			defer afterRequest()
			if server_time_offset, err = client.NewSetServerTimeService().Do(context.Background()); err == nil {
				err = &errorContinue{}
				client.TimeOffset = server_time_offset
			}
		}
	}
	return err
}
