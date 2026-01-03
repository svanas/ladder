package oneinch

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

type Extension struct {
	receiver      common.Address
	integratorFee IntegratorFee
	resolverFee   ResolverFee
}

func newExtension(receiver common.Address, integratorFee IntegratorFee, resolverFee ResolverFee) *Extension {
	return &Extension{receiver, integratorFee, resolverFee}
}

func trimPrefix(s, prefix string) string {
	if strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix)) {
		return s[len(prefix):]
	}
	return s
}

// encodeWhitelist encodes a list of Ethereum addresses into a single *big.Int value with a specific bit layout.
// returns an error if any address is invalid, has more than 255 addresses, or encounters issues during encoding.
func encodeWhitelist(whitelist []string) (*big.Int, error) {
	if len(whitelist) == 0 {
		return big.NewInt(0), nil
	}

	if len(whitelist) > 255 {
		return nil, fmt.Errorf("whitelist can have at most 255 addresses, got %d", len(whitelist))
	}

	encoded := big.NewInt(int64(len(whitelist)))

	mask80 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 80), big.NewInt(1))

	for _, address := range whitelist {
		address = trimPrefix(address, "0x")

		addressInt := new(big.Int)
		_, success := addressInt.SetString(address, 16)
		if !success {
			return nil, fmt.Errorf("invalid hex address: %s", address)
		}

		// mask to get only lower 80 bits
		addressLower80 := new(big.Int).And(addressInt, mask80)

		// shift encoded left by 80 bits and OR with the address
		encoded.Lsh(encoded, 80)
		encoded.Or(encoded, addressLower80)
	}

	return encoded, nil
}

// packFee encodes integrator and resolver fee details into a single 48-bit value with error checks on input ranges.
// returns the packed 48-bit value and an error if any input values exceed their allowed ranges.
func (e *Extension) packFee() (uint64, error) {
	integratorFee := uint64(e.integratorFee.Fee * 10)      // convert to basis points * 10
	integratorShare := uint64(e.integratorFee.Share / 100) // convert percentage to 0-100 range

	resolverFee := uint64(e.resolverFee.FeeBps * 10)                         // convert to basis points * 10
	resolverDiscount := uint64(100 - e.resolverFee.WhitelistDiscountPercent) // invert discount (100 - discount%)

	// range checks to ensure values fit in their allocated bit space
	if integratorFee > 0xffff {
		return 0, fmt.Errorf("integrator fee value must be between 0 and 65535, got %d", integratorFee)
	}
	if integratorShare > 0xff {
		return 0, fmt.Errorf("integrator share must be between 0 and 255, got %d", integratorShare)
	}
	if resolverFee > 0xffff {
		return 0, fmt.Errorf("resolver fee value must be between 0 and 65535, got %d", resolverFee)
	}
	if resolverDiscount > 0xff {
		return 0, fmt.Errorf("resolver discount must be between 0 and 255, got %d", resolverDiscount)
	}

	packed := (integratorFee << 32) | // bits 47-32 (16 bits)
		(integratorShare << 24) | // bits 31-24 (8 bits)
		(resolverFee << 8) | // bits 23-8 (16 bits)
		resolverDiscount // bits 7-0 (8 bits)

	return packed, nil
}

// concatFeeAndWhitelist combines fee parameters and whitelist into a single encoded value.
// returns the combined value and the total bit length.
func (e *Extension) concatFeeAndWhitelist(whitelist []string) (*big.Int, int, error) {
	packedFee, err := e.packFee()
	if err != nil {
		return nil, 0, err
	}

	encodedWhitelist, err := encodeWhitelist(whitelist)
	if err != nil {
		return nil, 0, err
	}

	// calculate whitelist bit length
	// if empty: 0 bits
	// otherwise: 8 bits (length) + (number of addresses * 80 bits each)
	var whitelistBitLen int
	if len(whitelist) == 0 {
		whitelistBitLen = 0
	} else {
		whitelistBitLen = 8 + (len(whitelist) * 80)
	}

	// packedFee << whitelistBitLen | encodedWhitelist
	feeAndWhitelist := new(big.Int).SetUint64(packedFee)
	feeAndWhitelist.Lsh(feeAndWhitelist, uint(whitelistBitLen))
	feeAndWhitelist.Or(feeAndWhitelist, encodedWhitelist)

	totalBitLen := 48 + whitelistBitLen

	return feeAndWhitelist, totalBitLen, nil
}

// getPostInteractionData encodes interaction data combining receiver, fee, whitelist, and optional interaction inputs into a byte array.
func (e *Extension) getPostInteractionData(whitelist []string) ([]byte, error) {
	interaction := big.NewInt(0)

	// add integrator address (20 bytes = 160 bits)
	interaction.Or(interaction, func(address string) *big.Int {
		integrator := new(big.Int)
		integrator.SetString(trimPrefix(address, "0x"), 16)
		return integrator
	}(e.integratorFee.Integrator))

	// add resolver/protocol fee receiver address (20 bytes = 160 bits)
	interaction.Lsh(interaction, 160)
	interaction.Or(interaction, func(address string) *big.Int {
		resolver := new(big.Int)
		resolver.SetString(trimPrefix(address, "0x"), 16)
		return resolver
	}(e.resolverFee.ProtocolFeeReceiver))

	// add receiver address (20 bytes = 160 bits)
	interaction.Lsh(interaction, 160)
	interaction.Or(interaction, func(address string) *big.Int {
		receiver := new(big.Int)
		receiver.SetString(trimPrefix(address, "0x"), 16)
		return receiver
	}(e.receiver.Hex()))

	// add fee and whitelist data
	feeAndWhitelist, bitLen, err := e.concatFeeAndWhitelist(whitelist)
	if err != nil {
		return nil, err
	}
	interaction.Lsh(interaction, uint(bitLen))
	interaction.Or(interaction, feeAndWhitelist)

	// calculate expected byte length
	expectedLen := 1 + 20 + 20 + 20 // flag + integrator + resolver + receiver
	if len(whitelist) == 0 {
		expectedLen += 6 // packedFee (48 bits = 6 bytes)
	} else {
		expectedLen += (48 + 8 + len(whitelist)*80) / 8 // fee + whitelist
	}

	interactionBytes := interaction.Bytes()

	result := make([]byte, expectedLen)
	result[0] = 0x01 // receiver
	copy(result[expectedLen-len(interactionBytes):], interactionBytes)

	return result, nil
}

// builds the order extension. returns the encoded extension as a hex string.
func (e *Extension) encode() (string, error) {

	var whitelist []string
	for _, value := range e.resolverFee.Whitelist {
		whitelist = append(whitelist, value)
		sort.Strings(whitelist) // sorting the final list to ensure a deterministic order
	}

	postInteractionData, err := e.getPostInteractionData(whitelist)
	if err != nil {
		return "", err
	}

	makingTakingAmountData, makingTakingAmountDataLen, err := e.concatFeeAndWhitelist(whitelist)
	if err != nil {
		return "", err
	}

	extensionTarget := make([]byte, 20)
	extensionAddress := func(address string) *big.Int {
		extensionAddress := new(big.Int)
		extensionAddress.SetString(trimPrefix(address, "0x"), 16)
		return extensionAddress
	}(e.resolverFee.ExtensionAddress).Bytes()
	copy(extensionTarget[20-len(extensionAddress):], extensionAddress)

	makingTaking := makingTakingAmountData.Bytes()
	expectedMakingTakingLen := (makingTakingAmountDataLen + 7) / 8 // round up to bytes
	if len(makingTaking) < expectedMakingTakingLen {
		padded := make([]byte, expectedMakingTakingLen)
		copy(padded[expectedMakingTakingLen-len(makingTaking):], makingTaking)
		makingTaking = padded
	}

	interactions := [][]byte{
		{},                                       // MakerAssetSuffix (empty)
		{},                                       // TakerAssetSuffix (empty)
		append(extensionTarget, makingTaking...), // MakingAmountData
		append(extensionTarget, makingTaking...), // TakingAmountData (same as making)
		{},                                       // Predicate (empty)
		{},                                       // MakerPermit (empty)
		{},                                       // PreInteractionData (empty)
		append(extensionTarget, postInteractionData...), // PostInteractionData
	}

	return encodeInteractions(interactions), nil
}

// encodeInteractions builds an extension hex string from byte slices
func encodeInteractions(interactions [][]byte) string {
	var byteCounts []int
	var dataBytes []byte

	// process first 8 interactions (these are used in the cumulative sum calculation)
	for i := 0; i < len(interactions) && i < 8; i++ {
		byteCounts = append(byteCounts, len(interactions[i]))
		dataBytes = append(dataBytes, interactions[i]...)
	}

	// add customData if present (no offset data)
	if len(interactions) > 8 {
		dataBytes = append(dataBytes, interactions[8]...)
	}

	// calculate cumulative offsets
	cumulativeSum := 0
	var offsets []byte
	for i := 0; i < len(byteCounts); i++ {
		cumulativeSum += byteCounts[i]
		offsetBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(offsetBytes, uint32(cumulativeSum))
		offsets = append(offsetBytes, offsets...)
	}

	// if no data, return an empty extension
	if len(dataBytes) == 0 {
		return "0x"
	}

	return "0x" + hex.EncodeToString(offsets) + hex.EncodeToString(dataBytes)
}
