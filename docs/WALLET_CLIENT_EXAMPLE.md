# Wallet Security Client Implementation

## JavaScript/TypeScript Example

### Wallet Security Manager Class

```typescript
class WalletSecurityManager {
  private sessionToken: string | null = null;
  private sessionExpiry: Date | null = null;
  private baseURL: string;
  private jwtToken: string;

  constructor(baseURL: string, jwtToken: string) {
    this.baseURL = baseURL;
    this.jwtToken = jwtToken;
  }

  /**
   * Create a new wallet session
   */
  async createSession(password: string): Promise<void> {
    try {
      const response = await fetch(`${this.baseURL}/api/v1/wallet/secure/session`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.jwtToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ password })
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Failed to create wallet session');
      }

      const data = await response.json();
      this.sessionToken = data.sessionToken;
      this.sessionExpiry = new Date(data.expiresAt);
      
      console.log('Wallet session created successfully');
    } catch (error) {
      console.error('Failed to create wallet session:', error);
      throw error;
    }
  }

  /**
   * Check if session is valid and not expired
   */
  isSessionValid(): boolean {
    if (!this.sessionToken || !this.sessionExpiry) {
      return false;
    }
    return new Date() < this.sessionExpiry;
  }

  /**
   * Get current nonce for wallet
   */
  async getNonce(): Promise<number> {
    try {
      const response = await fetch(`${this.baseURL}/api/v1/wallet/secure/nonce`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${this.jwtToken}`
        }
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Failed to get nonce');
      }

      const data = await response.json();
      return data.nonce;
    } catch (error) {
      console.error('Failed to get nonce:', error);
      throw error;
    }
  }

  /**
   * Sign a message using wallet session
   */
  async personalSign(message: string): Promise<string> {
    await this.ensureValidSession();

    try {
      const response = await fetch(`${this.baseURL}/api/v1/wallet/ops/personal-sign`, {
        method: 'POST',
        headers: {
          'X-Wallet-Session': this.sessionToken!,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ message })
      });

      if (!response.ok) {
        const error = await response.json();
        if (error.error === 'session expired') {
          this.sessionToken = null;
          this.sessionExpiry = null;
          throw new Error('Session expired. Please create a new session.');
        }
        throw new Error(error.error || 'Failed to sign message');
      }

      const data = await response.json();
      return data.signature;
    } catch (error) {
      console.error('Failed to sign message:', error);
      throw error;
    }
  }

  /**
   * Send a transaction using wallet session
   */
  async sendTransaction(transaction: {
    to: string;
    value: string;
    gasPrice: string;
    gas: number;
  }): Promise<string> {
    await this.ensureValidSession();

    // Get current nonce
    const nonce = await this.getNonce();

    const txData = {
      ...transaction,
      nonce
    };

    try {
      const response = await fetch(`${this.baseURL}/api/v1/wallet/ops/send-transaction`, {
        method: 'POST',
        headers: {
          'X-Wallet-Session': this.sessionToken!,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ transaction: txData })
      });

      if (!response.ok) {
        const error = await response.json();
        if (error.error === 'session expired') {
          this.sessionToken = null;
          this.sessionExpiry = null;
          throw new Error('Session expired. Please create a new session.');
        }
        if (error.error.includes('invalid nonce')) {
          throw new Error('Nonce mismatch. Please try again.');
        }
        throw new Error(error.error || 'Failed to send transaction');
      }

      const data = await response.json();
      return data.transactionHash;
    } catch (error) {
      console.error('Failed to send transaction:', error);
      throw error;
    }
  }

  /**
   * Sign typed data using wallet session
   */
  async signTypedData(typedData: any): Promise<string> {
    await this.ensureValidSession();

    try {
      const response = await fetch(`${this.baseURL}/api/v1/wallet/ops/sign-typed-data`, {
        method: 'POST',
        headers: {
          'X-Wallet-Session': this.sessionToken!,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ typedData })
      });

      if (!response.ok) {
        const error = await response.json();
        if (error.error === 'session expired') {
          this.sessionToken = null;
          this.sessionExpiry = null;
          throw new Error('Session expired. Please create a new session.');
        }
        throw new Error(error.error || 'Failed to sign typed data');
      }

      const data = await response.json();
      return data.signature;
    } catch (error) {
      console.error('Failed to sign typed data:', error);
      throw error;
    }
  }

  /**
   * Revoke current session
   */
  async revokeSession(): Promise<void> {
    if (!this.sessionToken) {
      return;
    }

    try {
      const response = await fetch(`${this.baseURL}/api/v1/wallet/secure/session`, {
        method: 'DELETE',
        headers: {
          'X-Wallet-Session': this.sessionToken
        }
      });

      if (response.ok) {
        console.log('Wallet session revoked successfully');
      }
    } catch (error) {
      console.error('Failed to revoke session:', error);
    } finally {
      this.sessionToken = null;
      this.sessionExpiry = null;
    }
  }

  /**
   * Get audit logs
   */
  async getAuditLogs(limit: number = 50): Promise<any[]> {
    try {
      const response = await fetch(`${this.baseURL}/api/v1/wallet/secure/audit-logs?limit=${limit}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${this.jwtToken}`
        }
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Failed to get audit logs');
      }

      const data = await response.json();
      return data.logs;
    } catch (error) {
      console.error('Failed to get audit logs:', error);
      throw error;
    }
  }

  /**
   * Ensure session is valid, create new one if needed
   */
  private async ensureValidSession(): Promise<void> {
    if (!this.isSessionValid()) {
      throw new Error('No valid session. Please create a new session with password.');
    }
  }
}
```

### Usage Examples

#### Basic Usage

```typescript
// Initialize wallet security manager
const walletManager = new WalletSecurityManager(
  'https://api.vybes.com',
  'your-jwt-token'
);

// Create session (user enters password)
async function setupWalletSession(password: string) {
  try {
    await walletManager.createSession(password);
    console.log('Wallet session ready');
  } catch (error) {
    console.error('Failed to setup wallet session:', error);
  }
}

// Sign a message
async function signMessage(message: string) {
  try {
    const signature = await walletManager.personalSign(message);
    console.log('Message signed:', signature);
    return signature;
  } catch (error) {
    console.error('Failed to sign message:', error);
    throw error;
  }
}

// Send a transaction
async function sendETH(toAddress: string, amount: string) {
  try {
    const txHash = await walletManager.sendTransaction({
      to: toAddress,
      value: amount,
      gasPrice: '20000000000', // 20 gwei
      gas: 21000
    });
    console.log('Transaction sent:', txHash);
    return txHash;
  } catch (error) {
    console.error('Failed to send transaction:', error);
    throw error;
  }
}
```

#### React Hook Example

```typescript
import { useState, useEffect, useCallback } from 'react';

interface UseWalletSecurityProps {
  baseURL: string;
  jwtToken: string;
}

export function useWalletSecurity({ baseURL, jwtToken }: UseWalletSecurityProps) {
  const [walletManager, setWalletManager] = useState<WalletSecurityManager | null>(null);
  const [isSessionValid, setIsSessionValid] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const manager = new WalletSecurityManager(baseURL, jwtToken);
    setWalletManager(manager);
  }, [baseURL, jwtToken]);

  const createSession = useCallback(async (password: string) => {
    if (!walletManager) return;

    setIsLoading(true);
    setError(null);

    try {
      await walletManager.createSession(password);
      setIsSessionValid(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create session');
      setIsSessionValid(false);
    } finally {
      setIsLoading(false);
    }
  }, [walletManager]);

  const signMessage = useCallback(async (message: string) => {
    if (!walletManager) throw new Error('Wallet manager not initialized');

    try {
      return await walletManager.personalSign(message);
    } catch (err) {
      if (err instanceof Error && err.message.includes('session expired')) {
        setIsSessionValid(false);
      }
      throw err;
    }
  }, [walletManager]);

  const sendTransaction = useCallback(async (transaction: any) => {
    if (!walletManager) throw new Error('Wallet manager not initialized');

    try {
      return await walletManager.sendTransaction(transaction);
    } catch (err) {
      if (err instanceof Error && err.message.includes('session expired')) {
        setIsSessionValid(false);
      }
      throw err;
    }
  }, [walletManager]);

  const revokeSession = useCallback(async () => {
    if (!walletManager) return;

    try {
      await walletManager.revokeSession();
      setIsSessionValid(false);
    } catch (err) {
      console.error('Failed to revoke session:', err);
    }
  }, [walletManager]);

  return {
    isSessionValid,
    isLoading,
    error,
    createSession,
    signMessage,
    sendTransaction,
    revokeSession
  };
}
```

#### React Component Example

```typescript
import React, { useState } from 'react';
import { useWalletSecurity } from './useWalletSecurity';

interface WalletComponentProps {
  baseURL: string;
  jwtToken: string;
}

export function WalletComponent({ baseURL, jwtToken }: WalletComponentProps) {
  const [password, setPassword] = useState('');
  const [message, setMessage] = useState('');
  const [toAddress, setToAddress] = useState('');
  const [amount, setAmount] = useState('');

  const {
    isSessionValid,
    isLoading,
    error,
    createSession,
    signMessage,
    sendTransaction,
    revokeSession
  } = useWalletSecurity({ baseURL, jwtToken });

  const handleCreateSession = async (e: React.FormEvent) => {
    e.preventDefault();
    await createSession(password);
  };

  const handleSignMessage = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const signature = await signMessage(message);
      alert(`Message signed: ${signature}`);
    } catch (err) {
      alert(`Failed to sign message: ${err}`);
    }
  };

  const handleSendTransaction = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const txHash = await sendTransaction({
        to: toAddress,
        value: amount,
        gasPrice: '20000000000',
        gas: 21000
      });
      alert(`Transaction sent: ${txHash}`);
    } catch (err) {
      alert(`Failed to send transaction: ${err}`);
    }
  };

  return (
    <div>
      <h2>Wallet Security</h2>
      
      {error && <div style={{ color: 'red' }}>Error: {error}</div>}
      
      {!isSessionValid ? (
        <form onSubmit={handleCreateSession}>
          <h3>Create Wallet Session</h3>
          <input
            type="password"
            placeholder="Enter password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          <button type="submit" disabled={isLoading}>
            {isLoading ? 'Creating Session...' : 'Create Session'}
          </button>
        </form>
      ) : (
        <div>
          <h3>Wallet Session Active</h3>
          <button onClick={revokeSession}>Revoke Session</button>
          
          <form onSubmit={handleSignMessage}>
            <h4>Sign Message</h4>
            <input
              type="text"
              placeholder="Enter message to sign"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              required
            />
            <button type="submit">Sign Message</button>
          </form>
          
          <form onSubmit={handleSendTransaction}>
            <h4>Send Transaction</h4>
            <input
              type="text"
              placeholder="To address"
              value={toAddress}
              onChange={(e) => setToAddress(e.target.value)}
              required
            />
            <input
              type="text"
              placeholder="Amount (wei)"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              required
            />
            <button type="submit">Send Transaction</button>
          </form>
        </div>
      )}
    </div>
  );
}
```

### Error Handling

```typescript
// Comprehensive error handling
class WalletError extends Error {
  constructor(
    message: string,
    public code: string,
    public retryable: boolean = false
  ) {
    super(message);
    this.name = 'WalletError';
  }
}

function handleWalletError(error: any): WalletError {
  if (error.message.includes('session expired')) {
    return new WalletError('Session expired', 'SESSION_EXPIRED', true);
  }
  
  if (error.message.includes('invalid nonce')) {
    return new WalletError('Nonce mismatch', 'NONCE_MISMATCH', true);
  }
  
  if (error.message.includes('rate limit')) {
    return new WalletError('Rate limit exceeded', 'RATE_LIMIT', true);
  }
  
  if (error.message.includes('account locked')) {
    return new WalletError('Account temporarily locked', 'ACCOUNT_LOCKED', false);
  }
  
  return new WalletError('Unknown wallet error', 'UNKNOWN', false);
}

// Retry mechanism
async function withRetry<T>(
  operation: () => Promise<T>,
  maxRetries: number = 3,
  delay: number = 1000
): Promise<T> {
  let lastError: Error;
  
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await operation();
    } catch (error) {
      lastError = error as Error;
      const walletError = handleWalletError(error);
      
      if (!walletError.retryable) {
        throw walletError;
      }
      
      if (i < maxRetries - 1) {
        await new Promise(resolve => setTimeout(resolve, delay * Math.pow(2, i)));
      }
    }
  }
  
  throw lastError!;
}
```

### Security Best Practices

```typescript
// Secure storage for session tokens
class SecureStorage {
  private static readonly SESSION_KEY = 'wallet_session_token';
  private static readonly EXPIRY_KEY = 'wallet_session_expiry';

  static saveSession(token: string, expiry: Date): void {
    // Use secure storage (keychain, secure enclave, etc.)
    if (typeof window !== 'undefined' && window.crypto) {
      // Browser: use sessionStorage for temporary storage
      sessionStorage.setItem(this.SESSION_KEY, token);
      sessionStorage.setItem(this.EXPIRY_KEY, expiry.toISOString());
    } else {
      // Node.js: use environment variables or secure key storage
      process.env[this.SESSION_KEY] = token;
      process.env[this.EXPIRY_KEY] = expiry.toISOString();
    }
  }

  static getSession(): { token: string; expiry: Date } | null {
    try {
      let token: string;
      let expiryStr: string;

      if (typeof window !== 'undefined' && window.crypto) {
        token = sessionStorage.getItem(this.SESSION_KEY) || '';
        expiryStr = sessionStorage.getItem(this.EXPIRY_KEY) || '';
      } else {
        token = process.env[this.SESSION_KEY] || '';
        expiryStr = process.env[this.EXPIRY_KEY] || '';
      }

      if (!token || !expiryStr) {
        return null;
      }

      const expiry = new Date(expiryStr);
      if (new Date() >= expiry) {
        this.clearSession();
        return null;
      }

      return { token, expiry };
    } catch (error) {
      this.clearSession();
      return null;
    }
  }

  static clearSession(): void {
    if (typeof window !== 'undefined' && window.crypto) {
      sessionStorage.removeItem(this.SESSION_KEY);
      sessionStorage.removeItem(this.EXPIRY_KEY);
    } else {
      delete process.env[this.SESSION_KEY];
      delete process.env[this.EXPIRY_KEY];
    }
  }
}
```

## Mobile App Example (React Native)

```typescript
import AsyncStorage from '@react-native-async-storage/async-storage';
import * as Keychain from 'react-native-keychain';

class MobileWalletSecurityManager extends WalletSecurityManager {
  private static readonly SESSION_KEY = 'wallet_session_token';
  private static readonly EXPIRY_KEY = 'wallet_session_expiry';

  async createSession(password: string): Promise<void> {
    await super.createSession(password);
    
    // Store session securely in keychain
    if (this.sessionToken && this.sessionExpiry) {
      await Keychain.setInternetCredentials(
        'wallet_session',
        'user',
        JSON.stringify({
          token: this.sessionToken,
          expiry: this.sessionExpiry.toISOString()
        })
      );
    }
  }

  async loadStoredSession(): Promise<boolean> {
    try {
      const credentials = await Keychain.getInternetCredentials('wallet_session');
      if (credentials && credentials.password) {
        const sessionData = JSON.parse(credentials.password);
        const expiry = new Date(sessionData.expiry);
        
        if (new Date() < expiry) {
          this.sessionToken = sessionData.token;
          this.sessionExpiry = expiry;
          return true;
        } else {
          await this.clearStoredSession();
        }
      }
    } catch (error) {
      console.error('Failed to load stored session:', error);
    }
    
    return false;
  }

  async clearStoredSession(): Promise<void> {
    try {
      await Keychain.resetInternetCredentials('wallet_session');
    } catch (error) {
      console.error('Failed to clear stored session:', error);
    }
  }

  async revokeSession(): Promise<void> {
    await super.revokeSession();
    await this.clearStoredSession();
  }
}
```

## Testing

```typescript
// Unit tests for wallet security manager
describe('WalletSecurityManager', () => {
  let walletManager: WalletSecurityManager;
  let mockFetch: jest.Mock;

  beforeEach(() => {
    mockFetch = jest.fn();
    global.fetch = mockFetch;
    walletManager = new WalletSecurityManager('https://api.test.com', 'test-jwt');
  });

  it('should create session successfully', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({
        sessionToken: 'test-token',
        expiresAt: '2024-01-01T12:00:00Z'
      })
    });

    await walletManager.createSession('password');
    
    expect(walletManager.isSessionValid()).toBe(true);
  });

  it('should handle session expiry', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      json: () => Promise.resolve({ error: 'session expired' })
    });

    await expect(walletManager.personalSign('test')).rejects.toThrow('Session expired');
  });

  it('should handle rate limiting', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      json: () => Promise.resolve({ error: 'Wallet operation rate limit exceeded' })
    });

    await expect(walletManager.personalSign('test')).rejects.toThrow('rate limit');
  });
});
```