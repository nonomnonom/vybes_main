# Vybes API - Railway Deployment Guide

## Architecture Overview

The Vybes application is designed as a microservices architecture with the following components:

### Core Services

1. **API Nodes** (Load Balanced)
   - **api-1** & **api-2**: Go-based REST API using Gin framework
   - Handles user authentication, content management, social features
   - Port: 8080
   - Health check endpoint: `/health`
   - Load balanced for high availability

2. **Worker Nodes** (MongoDB + Redis + NATS Cluster)
   - **worker-1**, **worker-2**, **worker-3**: Combined database, cache, and message queue
   - MongoDB 7.0 replica set for data persistence
   - Redis cluster for session management and caching
   - NATS cluster for real-time notifications
   - High availability with automatic failover

3. **External Storage** (Cloudflare R2)
   - **Cloudflare R2**: S3-compatible object storage
   - Stores posts, stories, user avatars, and media files
   - Global CDN for fast content delivery
   - No egress fees, cost-effective storage solution

## Railway Configuration

### railway.toml Structure

The `railway.toml` file defines:

- **Build Configuration**: Uses Dockerfile for building the Go application
- **Service Definitions**: Each component as a separate service
- **Environment Variables**: Configuration for each service
- **Dependencies**: Service startup order
- **Scaling**: Auto-scaling based on CPU/memory usage
- **Health Checks**: Monitoring service health
- **Resource Limits**: CPU and memory allocation

### Environment Variables Required

You need to set these environment variables in Railway:

#### Security & Authentication
- `JWT_SECRET`: Secret key for JWT token signing
- `WALLET_ENCRYPTION_KEY`: Key for encrypting wallet data

#### Database
- `MONGO_URI`: MongoDB connection string
- `MONGO_ROOT_USERNAME`: MongoDB root username
- `MONGO_ROOT_PASSWORD`: MongoDB root password

#### Cache
- `REDIS_PASSWORD`: Redis authentication password

#### External Storage (R2)
- `R2_ACCOUNT_ID`: Cloudflare R2 account ID
- `R2_ACCESS_KEY_ID`: R2 API access key ID
- `R2_SECRET_ACCESS_KEY`: R2 API secret access key
- `R2_ENDPOINT`: R2 endpoint URL
- `R2_POSTS_BUCKET`: Bucket name for posts
- `R2_STORIES_BUCKET`: Bucket name for stories

#### External Services
- `RESEND_API_KEY`: Resend email service API key
- `SENDER_EMAIL`: Email address for sending notifications
- `ETH_RPC_URL`: Ethereum RPC endpoint for blockchain features

#### Message Queue
- `NATS_URL`: NATS connection URL

## Deployment Steps

### 1. Prepare Your Repository

Ensure your repository contains:
- `railway.toml` (Railway configuration)
- `Dockerfile` (Application build instructions)
- `.railwayignore` (Files to exclude from deployment)
- All source code

### 2. Connect to Railway

1. Install Railway CLI:
   ```bash
   npm install -g @railway/cli
   ```

2. Login to Railway:
   ```bash
   railway login
   ```

3. Link your project:
   ```bash
   railway link
   ```

### 3. Set Environment Variables

Set all required environment variables in Railway dashboard or via CLI:

```bash
railway variables set JWT_SECRET=your-jwt-secret
railway variables set WALLET_ENCRYPTION_KEY=your-wallet-key
railway variables set MONGO_URI=mongodb://mongodb:27017
railway variables set REDIS_ADDR=redis:6379
railway variables set NATS_URL=nats://nats:4222
railway variables set R2_ACCOUNT_ID=your-r2-account-id
railway variables set R2_ACCESS_KEY_ID=your-r2-access-key-id
railway variables set R2_SECRET_ACCESS_KEY=your-r2-secret-access-key
railway variables set R2_ENDPOINT=https://your-account-id.r2.cloudflarestorage.com
railway variables set R2_POSTS_BUCKET=vybes-posts
railway variables set R2_STORIES_BUCKET=vybes-stories
# ... set all other required variables
```

### 4. Deploy

Deploy your application:

```bash
railway up
```

### 5. Monitor Deployment

Check deployment status:

```bash
railway status
```

View logs:

```bash
railway logs
```

## Service Communication

### Internal Networking

Services communicate using Railway's internal networking with cluster endpoints:

- **MongoDB Cluster**: `mongodb://worker-1:27017,worker-2:27017,worker-3:27017/vybes_production?replicaSet=vybes-rs`
- **Redis Cluster**: `worker-1:6379,worker-2:6379,worker-3:6379`
- **NATS Cluster**: `nats://worker-1:4222,nats://worker-2:4222,nats://worker-3:4222`

### External Storage (R2)

- **R2 Endpoint**: `https://{ACCOUNT_ID}.r2.cloudflarestorage.com`
- **Authentication**: API keys managed through Cloudflare dashboard
- **Buckets**: Separate buckets for posts and stories

### Health Checks

Each service includes health checks:
- API: `GET /health` returns 200 OK
- Database: Connection verification
- Cache: Redis ping
- Storage: R2 bucket access verification
- Message Queue: NATS connection

## Scaling Configuration

### Auto-scaling Rules

- **Minimum Replicas**: 1
- **Maximum Replicas**: 10
- **CPU Threshold**: 80%
- **Memory Threshold**: 85%

### Resource Allocation

- **CPU**: Auto-allocated by Railway based on demand
- **Memory**: Auto-allocated by Railway based on usage
- **No fixed limits**: Allows Railway to optimize performance

## Environment Management

### Development Environment
- Database: `vybes_development`
- Debug logging enabled

### Staging Environment
- Database: `vybes_staging`
- Production-like configuration

### Production Environment
- Database: `vybes_production`
- Full production configuration

## Monitoring & Logging

### Health Monitoring
- Railway automatically monitors service health
- Failed health checks trigger service restarts
- Maximum 10 restart attempts before marking as failed

### Logging
- Structured JSON logging via zerolog
- Log levels: DEBUG, INFO, WARN, ERROR, FATAL
- Railway provides centralized log viewing

## Troubleshooting

### Common Issues

1. **Service Startup Failures**
   - Check environment variables are set correctly
   - Verify service dependencies are running
   - Review logs for specific error messages

2. **Database Connection Issues**
   - Ensure MongoDB service is running
   - Verify connection string format
   - Check authentication credentials

3. **Storage Access Problems**
   - Verify R2 credentials in Cloudflare dashboard
   - Check bucket permissions and policies
   - Ensure R2 endpoint is correctly configured

4. **Message Queue Issues**
   - Confirm NATS service is running
   - Check connection URL format
   - Verify subscription permissions

### Debug Commands

```bash
# Check service status
railway status

# View service logs
railway logs --service vybes-api

# Connect to service shell
railway shell --service vybes-api

# Check environment variables
railway variables
```

## Security Considerations

### Environment Variables
- Never commit secrets to version control
- Use Railway's secure variable storage
- Rotate secrets regularly

### Network Security
- Services communicate over internal network
- External access only through API gateway
- SSL/TLS enabled for production

### Data Protection
- Database connections use authentication
- Redis requires password
- R2 uses IAM policies and bucket policies
- JWT tokens for API authentication

## Performance Optimization

### Caching Strategy
- Redis for session storage
- User profile caching
- Feed content caching
- API response caching

### Database Optimization
- Proper indexing on MongoDB collections
- Connection pooling
- Query optimization

### Storage Optimization
- R2 for efficient object storage
- Image compression for uploads
- Global CDN for fast content delivery
- No egress fees for cost optimization

## Backup & Recovery

### Database Backups
- MongoDB automated backups via Railway
- Point-in-time recovery available
- Cross-region backup replication

### Application Data
- R2 data replication and versioning
- Configuration version control
- Environment-specific backups

This deployment configuration provides a robust, scalable, and production-ready setup for the Vybes application on Railway's platform with cost-effective external storage via Cloudflare R2.