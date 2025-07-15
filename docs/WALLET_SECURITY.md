# Wallet Security System

## Overview

Sistem keamanan wallet yang baru telah diimplementasikan untuk menggantikan metode password-based yang berisiko. Sistem ini menggunakan session-based authentication dengan nonce management untuk transaksi Ethereum.

## Fitur Keamanan

### 1. Session-Based Wallet Access
- **Temporary Session Tokens**: Session token 32-byte yang aman secara kriptografis
- **Session Expiry**: Session otomatis expired setelah 30 menit
- **IP & User Agent Tracking**: Setiap session dicatat dengan IP dan User Agent
- **Session Revocation**: User dapat revoke session individual atau semua session

### 2. Nonce Management
- **Automatic Nonce Tracking**: Nonce otomatis di-track dan di-validate
- **Anti-Replay Protection**: Mencegah replay attack pada transaksi
- **Sequential Nonce**: Memastikan transaksi berurutan

### 3. Account Security
- **Failed Login Attempts**: Tracking percobaan login gagal
- **Account Locking**: Account otomatis di-lock setelah 4 percobaan gagal
- **Lock Duration**: 15 menit lock duration
- **Automatic Unlock**: Account otomatis unlock setelah waktu habis

### 4. Rate Limiting
- **Wallet Operations**: Max 10 operasi wallet per menit per user
- **Session Creation**: Rate limiting untuk pembuatan session
- **Configurable Limits**: Rate limit dapat dikonfigurasi

### 5. Audit Logging
- **Complete Audit Trail**: Semua operasi wallet di-log
- **Transaction Details**: Hash, nonce, amount, recipient address
- **Client Information**: IP address, User Agent
- **Status Tracking**: Success/error status untuk setiap operasi

## API Endpoints

### Session Management

#### Create Wallet Session
```http
POST /api/v1/wallet/secure/session
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "password": "user_password"
}
```

Response:
```json
{
  "sessionToken": "64_character_hex_string",
  "expiresAt": "2024-01-01T12:00:00Z",
  "message": "Wallet session created successfully"
}
```

#### Revoke Current Session
```http
DELETE /api/v1/wallet/secure/session
X-Wallet-Session: <session_token>
```

#### Revoke All Sessions
```http
DELETE /api/v1/wallet/secure/sessions
Authorization: Bearer <jwt_token>
```

#### Get Next Nonce
```http
GET /api/v1/wallet/secure/nonce
Authorization: Bearer <jwt_token>
```

Response:
```json
{
  "nonce": 5
}
```

#### Get Audit Logs
```http
GET /api/v1/wallet/secure/audit-logs?limit=50
Authorization: Bearer <jwt_token>
```

### Wallet Operations (Session-Based)

Semua operasi wallet sekarang menggunakan session token di header `X-Wallet-Session`:

#### Personal Sign
```http
POST /api/v1/wallet/ops/personal-sign
X-Wallet-Session: <session_token>
Content-Type: application/json

{
  "message": "Hello World"
}
```

#### Sign Transaction
```http
POST /api/v1/wallet/ops/sign-transaction
X-Wallet-Session: <session_token>
Content-Type: application/json

{
  "transaction": {
    "to": "0x...",
    "value": "0x...",
    "nonce": 5,
    "gasPrice": "0x...",
    "gas": 21000
  }
}
```

#### Send Transaction
```http
POST /api/v1/wallet/ops/send-transaction
X-Wallet-Session: <session_token>
Content-Type: application/json

{
  "transaction": {
    "to": "0x...",
    "value": "0x...",
    "nonce": 5,
    "gasPrice": "0x...",
    "gas": 21000
  }
}
```

#### Sign Typed Data V4
```http
POST /api/v1/wallet/ops/sign-typed-data
X-Wallet-Session: <session_token>
Content-Type: application/json

{
  "typedData": {
    "types": {...},
    "primaryType": "Person",
    "domain": {...},
    "message": {...}
  }
}
```

#### Secp256k1 Sign
```http
POST /api/v1/wallet/ops/secp256k1-sign
X-Wallet-Session: <session_token>
Content-Type: application/json

{
  "hash": "0x..."
}
```

## Migration Guide

### Dari Password-Based ke Session-Based

1. **Ganti endpoint wallet operations**:
   - Dari: `/api/v1/wallet/personal-sign`
   - Ke: `/api/v1/wallet/ops/personal-sign`

2. **Tambahkan session management**:
   ```javascript
   // 1. Create session
   const sessionResponse = await fetch('/api/v1/wallet/secure/session', {
     method: 'POST',
     headers: {
       'Authorization': `Bearer ${jwtToken}`,
       'Content-Type': 'application/json'
     },
     body: JSON.stringify({ password: userPassword })
   });
   
   const { sessionToken } = await sessionResponse.json();
   
   // 2. Use session for wallet operations
   const signature = await fetch('/api/v1/wallet/ops/personal-sign', {
     method: 'POST',
     headers: {
       'X-Wallet-Session': sessionToken,
       'Content-Type': 'application/json'
     },
     body: JSON.stringify({ message: 'Hello World' })
   });
   ```

3. **Handle session expiry**:
   ```javascript
   // Check if session expired and recreate if needed
   if (response.status === 401 && response.error === 'session expired') {
     // Recreate session
     await createWalletSession();
   }
   ```

## Security Best Practices

### 1. Session Management
- **Store session tokens securely**: Gunakan secure storage (keychain, secure enclave)
- **Auto-refresh sessions**: Refresh session sebelum expired
- **Revoke unused sessions**: Revoke session saat user logout
- **Monitor active sessions**: Track semua active sessions

### 2. Nonce Management
- **Always get current nonce**: Selalu ambil nonce terbaru sebelum transaksi
- **Handle nonce conflicts**: Handle kasus nonce tidak sesuai
- **Retry mechanism**: Implement retry untuk transaksi gagal

### 3. Error Handling
- **Graceful degradation**: Handle session expiry dengan baik
- **User feedback**: Berikan feedback yang jelas ke user
- **Logging**: Log semua error untuk debugging

### 4. Rate Limiting
- **Respect rate limits**: Handle rate limit responses
- **Exponential backoff**: Implement backoff untuk retry
- **User notification**: Beritahu user saat rate limit tercapai

## Database Schema

### New Collections

#### wallet_sessions
```javascript
{
  _id: ObjectId,
  userId: ObjectId,
  sessionToken: String,
  expiresAt: Date,
  ipAddress: String,
  userAgent: String,
  createdAt: Date,
  lastUsedAt: Date
}
```

#### wallet_audit_logs
```javascript
{
  _id: ObjectId,
  userId: ObjectId,
  action: String,
  txHash: String,
  nonce: Number,
  amount: String,
  toAddress: String,
  ipAddress: String,
  userAgent: String,
  status: String,
  errorMsg: String,
  createdAt: Date
}
```

### Updated User Collection
```javascript
{
  // ... existing fields
  walletNonce: Number,
  lastWalletAccess: Date,
  failedLoginAttempts: Number,
  accountLockedUntil: Date
}
```

## Monitoring & Alerts

### Key Metrics
- **Session creation rate**: Monitor pembuatan session
- **Failed login attempts**: Alert untuk multiple failed attempts
- **Rate limit violations**: Monitor rate limit hits
- **Audit log volume**: Monitor volume audit logs

### Alerts
- **Account lockouts**: Alert saat account di-lock
- **Suspicious activity**: Alert untuk aktivitas mencurigakan
- **Session anomalies**: Alert untuk session yang tidak normal

## Troubleshooting

### Common Issues

1. **Session Expired**
   - Error: "session expired"
   - Solution: Recreate session dengan password

2. **Invalid Nonce**
   - Error: "invalid nonce. Expected X, got Y"
   - Solution: Get current nonce dan retry

3. **Rate Limit Exceeded**
   - Error: "Wallet operation rate limit exceeded"
   - Solution: Wait dan retry setelah 1 menit

4. **Account Locked**
   - Error: "account is locked until ..."
   - Solution: Wait sampai lock period selesai

### Debug Information
- **Session ID**: Track session untuk debugging
- **Nonce values**: Log nonce untuk troubleshooting
- **Client info**: IP dan User Agent untuk security analysis
- **Audit logs**: Complete audit trail untuk investigation