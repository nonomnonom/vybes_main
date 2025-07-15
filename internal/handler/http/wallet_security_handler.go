package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"vybes/internal/domain"
	"vybes/internal/service"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gin-gonic/gin"
)

// WalletSecurityHandler handles HTTP requests for wallet security operations.
type WalletSecurityHandler struct {
	walletSecurityService service.WalletSecurityService
}

// NewWalletSecurityHandler creates a new WalletSecurityHandler.
func NewWalletSecurityHandler(walletSecurityService service.WalletSecurityService) *WalletSecurityHandler {
	return &WalletSecurityHandler{
		walletSecurityService: walletSecurityService,
	}
}

// CreateWalletSession creates a new wallet session.
func (h *WalletSecurityHandler) CreateWalletSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	var request struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	session, err := h.walletSecurityService.CreateWalletSession(
		c.Request.Context(),
		userID.(string),
		request.Password,
		ipAddress,
		userAgent,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessionToken": session.SessionToken,
		"expiresAt":    session.ExpiresAt,
		"message":      "Wallet session created successfully",
	})
}

// RevokeWalletSession revokes the current wallet session.
func (h *WalletSecurityHandler) RevokeWalletSession(c *gin.Context) {
	session, exists := c.Get("walletSession")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wallet session not found"})
		return
	}

	sessionToken := session.(*domain.WalletSession).SessionToken
	err := h.walletSecurityService.RevokeWalletSession(c.Request.Context(), sessionToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Wallet session revoked successfully"})
}

// RevokeAllSessions revokes all wallet sessions for the user.
func (h *WalletSecurityHandler) RevokeAllSessions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	err := h.walletSecurityService.RevokeAllUserSessions(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All wallet sessions revoked successfully"})
}

// GetNextNonce gets the next nonce for the user's wallet.
func (h *WalletSecurityHandler) GetNextNonce(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	nonce, err := h.walletSecurityService.GetNextNonce(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"nonce": nonce})
}

// GetWalletAuditLogs gets audit logs for the user's wallet operations.
func (h *WalletSecurityHandler) GetWalletAuditLogs(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	if limit > 100 {
		limit = 100 // Cap at 100 records
	}

	logs, err := h.walletSecurityService.GetWalletAuditLogs(c.Request.Context(), userID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// PersonalSignWithSession signs a message using wallet session.
func (h *WalletSecurityHandler) PersonalSignWithSession(c *gin.Context) {
	var request struct {
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, exists := c.Get("walletSession")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wallet session not found"})
		return
	}

	sessionToken := session.(*domain.WalletSession).SessionToken
	signature, err := h.walletSecurityService.PersonalSignWithSession(
		c.Request.Context(),
		sessionToken,
		request.Message,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"signature": signature})
}

// SignTransactionWithSession signs a transaction using wallet session.
func (h *WalletSecurityHandler) SignTransactionWithSession(c *gin.Context) {
	var request struct {
		Transaction json.RawMessage `json:"transaction" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var tx types.Transaction
	if err := json.Unmarshal(request.Transaction, &tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction format"})
		return
	}

	session, exists := c.Get("walletSession")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wallet session not found"})
		return
	}

	sessionToken := session.(*domain.WalletSession).SessionToken
	signedTx, err := h.walletSecurityService.SignTransactionWithSession(
		c.Request.Context(),
		sessionToken,
		&tx,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, signedTx)
}

// SendTransactionWithSession sends a transaction using wallet session.
func (h *WalletSecurityHandler) SendTransactionWithSession(c *gin.Context) {
	var request struct {
		Transaction json.RawMessage `json:"transaction" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var tx types.Transaction
	if err := json.Unmarshal(request.Transaction, &tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction format"})
		return
	}

	session, exists := c.Get("walletSession")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wallet session not found"})
		return
	}

	sessionToken := session.(*domain.WalletSession).SessionToken
	txHash, err := h.walletSecurityService.SendTransactionWithSession(
		c.Request.Context(),
		sessionToken,
		&tx,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactionHash": txHash.Hex()})
}

// SignTypedDataV4WithSession signs typed data using wallet session.
func (h *WalletSecurityHandler) SignTypedDataV4WithSession(c *gin.Context) {
	var request struct {
		TypedData apitypes.TypedData `json:"typedData" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, exists := c.Get("walletSession")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wallet session not found"})
		return
	}

	sessionToken := session.(*domain.WalletSession).SessionToken
	signature, err := h.walletSecurityService.SignTypedDataV4WithSession(
		c.Request.Context(),
		sessionToken,
		request.TypedData,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"signature": signature})
}

// Secp256k1SignWithSession signs a hash using wallet session.
func (h *WalletSecurityHandler) Secp256k1SignWithSession(c *gin.Context) {
	var request struct {
		Hash string `json:"hash" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, exists := c.Get("walletSession")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wallet session not found"})
		return
	}

	sessionToken := session.(*domain.WalletSession).SessionToken
	signature, err := h.walletSecurityService.Secp256k1SignWithSession(
		c.Request.Context(),
		sessionToken,
		request.Hash,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"signature": signature})
}