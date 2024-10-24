package oneinch

import (
	"fmt"
	"math/big"
)

const (
	noPartialFillsFlag      = 255 // if set, the order does not allow partial fills.
	allowMultipleFillsFlag  = 254 // if set, the order permits multiple fills.
	needPreinteractionFlag  = 252 // if set, the order requires pre-interaction call.
	needPostinteractionFlag = 251 // if set, the order requires post-interaction call.
	needEpochCheckFlag      = 250 // if set, an order uses epoch manager for cancelling.
	hasExtensionFlag        = 249 // if set, the order applies extension(s) logic during fill.
	usePermit2Flag          = 248 // if set, the order uses Uniswap Permit2.
	unwrapWethFlag          = 247 // if set, the order requires unwrapping WETH.
)

type MakerTraits struct {
	AllowedSender string
	Expiry        int64
	Nonce         int64
	Series        int64

	NeedPostinteraction bool
	NeedPreinteraction  bool
	NeedEpochCheck      bool
	HasExtension        bool
	ShouldUsePermit2    bool
	ShouldUnwrapWeth    bool

	AllowPartialFills  bool
	AllowMultipleFills bool
}

func newMakerTraits(epoch big.Int, expiry int64) *MakerTraits {
	return &MakerTraits{
		AllowedSender: "0x0000000000000000000000000000000000000000",
		Expiry:        expiry,
		Nonce:         epoch.Int64(),
		Series:        0,

		NeedPostinteraction: false,
		NeedPreinteraction:  false,
		NeedEpochCheck:      true,
		HasExtension:        false,
		ShouldUsePermit2:    false,
		ShouldUnwrapWeth:    false,

		AllowPartialFills:  true,
		AllowMultipleFills: true,
	}
}

func (mt *MakerTraits) encode() string {
	encodedCalldata := new(big.Int)

	tmp := new(big.Int)
	// limit orders require this flag to always be present
	if mt.AllowMultipleFills {
		encodedCalldata.Or(encodedCalldata, tmp.Lsh(big.NewInt(1), allowMultipleFillsFlag))
	}
	if mt.NeedPostinteraction {
		encodedCalldata.Or(encodedCalldata, tmp.Lsh(big.NewInt(1), needPostinteractionFlag))
	}
	if !mt.AllowPartialFills {
		encodedCalldata.Or(encodedCalldata, tmp.Lsh(big.NewInt(1), noPartialFillsFlag))
	}
	if mt.NeedPreinteraction {
		encodedCalldata.Or(encodedCalldata, tmp.Lsh(big.NewInt(1), needPreinteractionFlag))
	}
	if mt.NeedEpochCheck {
		encodedCalldata.Or(encodedCalldata, tmp.Lsh(big.NewInt(1), needEpochCheckFlag))
	}
	if mt.HasExtension {
		encodedCalldata.Or(encodedCalldata, tmp.Lsh(big.NewInt(1), hasExtensionFlag))
	}
	if mt.ShouldUsePermit2 {
		encodedCalldata.Or(encodedCalldata, tmp.Lsh(big.NewInt(1), usePermit2Flag))
	}
	if mt.ShouldUnwrapWeth {
		encodedCalldata.Or(encodedCalldata, tmp.Lsh(big.NewInt(1), unwrapWethFlag))
	}

	encodedCalldata.Or(encodedCalldata, tmp.Lsh(big.NewInt(mt.Series), 160))
	encodedCalldata.Or(encodedCalldata, tmp.Lsh(big.NewInt(mt.Nonce), 120))
	encodedCalldata.Or(encodedCalldata, tmp.Lsh(big.NewInt(mt.Expiry), 80))

	// convert AllowedSender from hex string to big.Int
	if mt.AllowedSender != "" {
		allowedSenderInt := new(big.Int)
		allowedSenderInt.SetString(mt.AllowedSender[len(mt.AllowedSender)-20:], 16) // we only care about the last 20 characters of the ethereum address
		encodedCalldata.Or(encodedCalldata, tmp.And(allowedSenderInt, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 80), big.NewInt(1))))
	}

	// pad the predicate to 32 bytes with 0's on the left and convert to hex string
	paddedPredicate := fmt.Sprintf("%032x", encodedCalldata)
	return "0x" + paddedPredicate
}
