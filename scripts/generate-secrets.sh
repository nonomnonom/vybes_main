#!/bin/bash

# =============================================================================
# VYBES API - SECRET GENERATOR SCRIPT
# =============================================================================
# This script generates secure secrets for Railway deployment
# Run this script to generate all required secrets

echo "🔐 VYBES API - Secret Generator"
echo "=================================="
echo ""

# Check if openssl is available
if ! command -v openssl &> /dev/null; then
    echo "❌ Error: openssl is not installed. Please install it first."
    exit 1
fi

echo "📝 Generating secure secrets for Railway deployment..."
echo ""

# Generate JWT Secret (32 bytes = 256 bits)
echo "🔑 JWT_SECRET:"
JWT_SECRET=$(openssl rand -base64 32)
echo "JWT_SECRET=$JWT_SECRET"
echo ""

# Generate Wallet Encryption Key (32 bytes = 256 bits)
echo "🔐 WALLET_ENCRYPTION_KEY:"
WALLET_ENCRYPTION_KEY=$(openssl rand -base64 32)
echo "WALLET_ENCRYPTION_KEY=$WALLET_ENCRYPTION_KEY"
echo ""

# Generate MongoDB Root Password (16 bytes = 128 bits)
echo "🗄️  MONGO_ROOT_PASSWORD:"
MONGO_ROOT_PASSWORD=$(openssl rand -base64 16)
echo "MONGO_ROOT_PASSWORD=$MONGO_ROOT_PASSWORD"
echo ""

# Generate Redis Password (16 bytes = 128 bits)
echo "📦 REDIS_PASSWORD:"
REDIS_PASSWORD=$(openssl rand -base64 16)
echo "REDIS_PASSWORD=$REDIS_PASSWORD"
echo ""

echo "✅ All secrets generated successfully!"
echo ""
echo "📋 Next steps:"
echo "1. Copy these values to Railway dashboard under Variables"
echo "2. Set the following additional variables manually:"
echo "   - RESEND_API_KEY (get from https://resend.com)"
echo "   - SENDER_EMAIL (your verified email domain)"
echo "   - ETH_RPC_URL (get from Infura, Alchemy, or Ankr)"
echo "   - MONGO_ROOT_USERNAME (e.g., 'admin')"
echo ""
echo "☁️  R2 Configuration (Cloudflare R2):"
echo "1. Go to Cloudflare Dashboard > R2 Object Storage"
echo "2. Create buckets: 'vybes-posts' and 'vybes-stories'"
echo "3. Go to Manage R2 API tokens"
echo "4. Create a new API token with appropriate permissions"
echo "5. Set these variables in Railway:"
echo "   - R2_ACCOUNT_ID (found in Cloudflare dashboard)"
echo "   - R2_ACCESS_KEY_ID (from API token)"
echo "   - R2_SECRET_ACCESS_KEY (from API token)"
echo "   - R2_ENDPOINT (https://{ACCOUNT_ID}.r2.cloudflarestorage.com)"
echo ""
echo "🔒 Security Notes:"
echo "- Keep these secrets secure and never commit them to version control"
echo "- Use different secrets for development, staging, and production"
echo "- Rotate secrets regularly in production"
echo "- R2 credentials are managed through Cloudflare dashboard"
echo ""
echo "🚀 Ready to deploy on Railway!"