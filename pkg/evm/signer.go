package evm

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// Signer holds the private key for signing operations.
type Signer struct {
	privateKey *ecdsa.PrivateKey
}

// NewSignerFromHex creates a new Signer from a hex-encoded private key.
func NewSignerFromHex(hexKey string) (*Signer, error) {
	privateKey, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return nil, fmt.Errorf("could not reconstruct private key: %w", err)
	}
	return &Signer{privateKey: privateKey}, nil
}

// PersonalSign signs a message using the personal_sign method.
func (s *Signer) PersonalSign(message []byte) (string, error) {
	fullMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(fullMessage))

	signature, err := crypto.Sign(hash.Bytes(), s.privateKey)
	if err != nil {
		return "", err
	}
	signature[64] += 27 // Adjust V value

	return hexutil.Encode(signature), nil
}

// SignTransaction signs an Ethereum transaction.
func (s *Signer) SignTransaction(tx *types.Transaction) (*types.Transaction, error) {
	chainID, err := s.getChainIDFromTx(tx)
	if err != nil {
		return nil, err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), s.privateKey)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

// SignTypedDataV4 signs a structured message according to EIP-712.
func (s *Signer) SignTypedDataV4(typedData apitypes.TypedData) (string, error) {
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return "", err
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return "", err
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	hash := crypto.Keccak256Hash(rawData)

	signature, err := crypto.Sign(hash.Bytes(), s.privateKey)
	if err != nil {
		return "", err
	}
	signature[64] += 27 // Adjust V value

	return hexutil.Encode(signature), nil
}

// Secp256k1Sign signs a raw hash.
func (s *Signer) Secp256k1Sign(hash []byte) (string, error) {
	signature, err := crypto.Sign(hash, s.privateKey)
	if err != nil {
		return "", err
	}
	signature[64] += 27 // Adjust V value

	return hexutil.Encode(signature), nil
}


// Helper function to get chain ID from transaction.
func (s *Signer) getChainIDFromTx(tx *types.Transaction) (*big.Int, error) {
	if tx.ChainId() == nil {
		return nil, fmt.Errorf("transaction does not have a chain id")
	}
	return tx.ChainId(), nil
}