package evm

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// Signer encapsulates the private key and provides methods for signing
// various types of Ethereum transactions and messages.
type Signer struct {
	privateKey *ecdsa.PrivateKey
}

// NewSignerFromHex creates a new Signer instance from a hex-encoded private key.
// The private key should be a 64-character hex string (32 bytes).
//
// Parameters:
//   - privateKeyHex: Hex-encoded private key string
//
// Returns:
//   - *Signer: A configured signer instance
//   - error: Any error that occurred during initialization
func NewSignerFromHex(privateKeyHex string) (*Signer, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}
	return &Signer{privateKey: privateKey}, nil
}

// PersonalSign signs a message using the Ethereum personal_sign method.
// This is commonly used for signing login messages and other user authentication.
//
// Parameters:
//   - message: The message to sign (will be prefixed with Ethereum signature prefix)
//
// Returns:
//   - string: Hex-encoded signature
//   - error: Any error that occurred during signing
func (s *Signer) PersonalSign(message string) (string, error) {
	hash := crypto.Keccak256Hash([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	signature, err := crypto.Sign(hash.Bytes(), s.privateKey)
	if err != nil {
		return "", err
	}
	signature[64] += 27 // Adjust V value for Ethereum signature format
	return hexutil.Encode(signature), nil
}

// SignTransaction signs an Ethereum transaction with the provided parameters.
// The transaction is signed using EIP-155 replay protection.
//
// Parameters:
//   - nonce: Transaction nonce
//   - to: Recipient address (nil for contract creation)
//   - value: Amount of ETH to send
//   - gasLimit: Maximum gas to use
//   - gasPrice: Gas price in wei
//   - data: Transaction data (nil for simple transfers)
//   - chainID: Network chain ID for replay protection
//
// Returns:
//   - string: Hex-encoded signed transaction
//   - error: Any error that occurred during signing
func (s *Signer) SignTransaction(nonce uint64, to *common.Address, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, chainID *big.Int) (string, error) {
	tx := types.NewTransaction(nonce, *to, value, gasLimit, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(signedTx.Hash().Bytes()), nil
}

// SignTypedDataV4 signs structured data according to EIP-712 specification.
// This is used for signing typed data with domain separation for security.
//
// Parameters:
//   - typedData: The structured data to sign according to EIP-712 format
//
// Returns:
//   - string: Hex-encoded signature
//   - error: Any error that occurred during signing
func (s *Signer) SignTypedDataV4(typedData apitypes.TypedData) (string, error) {
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return "", err
	}
	dataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return "", err
	}
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(dataHash)))
	hash := crypto.Keccak256Hash(rawData)
	signature, err := crypto.Sign(hash.Bytes(), s.privateKey)
	if err != nil {
		return "", err
	}
	signature[64] += 27 // Adjust V value for Ethereum signature format
	return hexutil.Encode(signature), nil
}

// Secp256k1Sign signs a raw hash using the secp256k1 curve.
// This is a low-level signing function for custom hash signing.
//
// Parameters:
//   - hash: The raw hash bytes to sign
//
// Returns:
//   - string: Hex-encoded signature
//   - error: Any error that occurred during signing
func (s *Signer) Secp256k1Sign(hash []byte) (string, error) {
	signature, err := crypto.Sign(hash, s.privateKey)
	if err != nil {
		return "", err
	}
	signature[64] += 27 // Adjust V value for Ethereum signature format
	return hexutil.Encode(signature), nil
}

// Helper function to extract chain ID from transaction for replay protection
func getChainID(tx *types.Transaction) *big.Int {
	if tx.ChainId() != nil {
		return tx.ChainId()
	}
	return big.NewInt(1) // Default to mainnet if not specified
}
