package service

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"vybes/internal/config"
	"vybes/pkg/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
)

// WalletService defines the interface for wallet operations.
type WalletService interface {
	CreateWallet() (address string, encryptedPrivateKey string, err error)
	SendTransaction(ctx context.Context, tx *types.Transaction, privateKeyHex string) (common.Hash, error)
}

type walletService struct {
	encryptionKey string
	ethClient     *ethclient.Client
}

// NewWalletService creates a new wallet service.
func NewWalletService(cfg *config.Config) WalletService {
	client, err := ethclient.Dial(cfg.EthRPCURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to the Ethereum client")
	}

	return &walletService{
		encryptionKey: cfg.WalletEncryptionKey,
		ethClient:     client,
	}
}

// CreateWallet generates a new Ethereum wallet and returns the address and encrypted private key.
func (s *walletService) CreateWallet() (string, string, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return "", "", err
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hexutil.Encode(privateKeyBytes)[2:] // remove 0x prefix

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", "", errors.New("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	encryptedPrivateKey, err := utils.Encrypt(privateKeyHex, s.encryptionKey)
	if err != nil {
		return "", "", err
	}

	return address, encryptedPrivateKey, nil
}

// SendTransaction signs and sends a transaction to the network.
func (s *walletService) SendTransaction(ctx context.Context, tx *types.Transaction, privateKeyHex string) (common.Hash, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return common.Hash{}, err
	}

	chainID, err := s.ethClient.ChainID(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return common.Hash{}, err
	}

	err = s.ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return common.Hash{}, err
	}

	return signedTx.Hash(), nil
}
