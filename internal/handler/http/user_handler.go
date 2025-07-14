package http

import (
	"encoding/json"
	"net/http"
	
	"vybes/internal/service"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for users.
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// ... (all other handlers remain the same)
func (h *UserHandler) Register(c *gin.Context) {
	var request struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.userService.Register(c.Request.Context(), request.Name, request.Email, request.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}
func (h *UserHandler) Login(c *gin.Context) {
	var request struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := h.userService.Login(c.Request.Context(), request.Email, request.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	viewerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}
	username := c.Param("username")
	profile, err := h.userService.GetUserProfile(c.Request.Context(), viewerID.(string), username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}
	var payload service.UpdateProfilePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedUser, err := h.userService.UpdateProfile(c.Request.Context(), userID.(string), payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedUser)
}
func (h *UserHandler) ExportPrivateKey(c *gin.Context) {
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
	privateKey, err := h.userService.ExportPrivateKey(c.Request.Context(), userID.(string), request.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"privateKey": privateKey})
}
func (h *UserHandler) PersonalSign(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}
	var request struct {
		Password string `json:"password" binding:"required"`
		Message  string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	signature, err := h.userService.PersonalSign(c.Request.Context(), userID.(string), request.Password, request.Message)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"signature": signature})
}
func (h *UserHandler) SignTransaction(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}
	var request struct {
		Password    string          `json:"password" binding:"required"`
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
	signedTx, err := h.userService.SignTransaction(c.Request.Context(), userID.(string), request.Password, &tx)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, signedTx)
}
func (h *UserHandler) SignTypedDataV4(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}
	var request struct {
		Password  string            `json:"password" binding:"required"`
		TypedData apitypes.TypedData `json:"typedData" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	signature, err := h.userService.SignTypedDataV4(c.Request.Context(), userID.(string), request.Password, request.TypedData)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"signature": signature})
}
func (h *UserHandler) SendTransaction(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}
	var request struct {
		Password    string          `json:"password" binding:"required"`
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
	txHash, err := h.userService.SendTransaction(c.Request.Context(), userID.(string), request.Password, &tx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transactionHash": txHash.Hex()})
}
func (h *UserHandler) Secp256k1Sign(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}
	var request struct {
		Password string `json:"password" binding:"required"`
		Hash     string `json:"hash" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	signature, err := h.userService.Secp256k1Sign(c.Request.Context(), userID.(string), request.Password, request.Hash)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"signature": signature})
}
func (h *UserHandler) RequestOTP(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.userService.RequestOTP(c.Request.Context(), request.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "If a user with that email exists, an OTP has been sent."})
}
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var request struct {
		Email       string `json:"email" binding:"required,email"`
		OTP         string `json:"otp" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.userService.VerifyOTPAndResetPassword(c.Request.Context(), request.Email, request.OTP, request.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully."})
}