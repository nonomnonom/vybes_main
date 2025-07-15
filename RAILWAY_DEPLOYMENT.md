# Vybes API - Railway Deployment Guide

## Architecture Overview

The Vybes application is designed as a microservices architecture with the following components:

### Core Services

1. **vybes-api** (Main Application)
   - Go-based REST API using Gin framework
   - Handles user authentication, content management, social features
   - Port: 8080
   - Health check endpoint: `/health`

2. **mongodb** (Database)
   - MongoDB 7.0 for data persistence
   - Stores user profiles, posts, stories, notifications, etc.
   - Port: 27017

3. **redis** (Cache)
   - Redis 7.2 for session management and caching
   - Improves performance for frequently accessed data
   - Port: 6379

4. **minio** (Object Storage)
   - MinIO for file storage (posts, stories, user avatars)
   - S3-compatible object storage
   - Ports: 9000 (API), 9001 (Console)

5. **nats** (Message Queue)
   - NATS 2.10 for real-time notifications
   - Handles async notification processing
   - Ports: 4222 (Client), 8222 (Monitoring)

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

#### Storage
- `MINIO_ACCESS_KEY`: MinIO access key
- `MINIO_SECRET_KEY`: MinIO secret key

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
railway variables set MINIO_ENDPOINT=minio:9000
railway variables set NATS_URL=nats://nats:4222
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

Services communicate using Railway's internal networking:

- **MongoDB**: `mongodb://mongodb:27017`
- **Redis**: `redis:6379`
- **MinIO**: `minio:9000`
- **NATS**: `nats://nats:4222`

### Health Checks

Each service includes health checks:
- API: `GET /health` returns 200 OK
- Database: Connection verification
- Cache: Redis ping
- Storage: MinIO bucket access
- Message Queue: NATS connection

## Scaling Configuration

### Auto-scaling Rules

- **Minimum Replicas**: 1
- **Maximum Replicas**: 5
- **CPU Threshold**: 70%
- **Memory Threshold**: 80%

### Resource Allocation

- **CPU**: 1000m (1 core)
- **Memory**: 2Gi

## Environment Management

### Development Environment
- Database: `vybes_development`
- MinIO SSL: Disabled
- Debug logging enabled

### Staging Environment
- Database: `vybes_staging`
- MinIO SSL: Enabled
- Production-like configuration

### Production Environment
- Database: `vybes_production`
- MinIO SSL: Enabled
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
   - Verify MinIO credentials
   - Check bucket permissions
   - Ensure SSL configuration matches environment

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
- MinIO uses access keys
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
- MinIO for efficient object storage
- Image compression for uploads
- CDN integration for static assets

## Backup & Recovery

### Database Backups
- MongoDB automated backups via Railway
- Point-in-time recovery available
- Cross-region backup replication

### Application Data
- MinIO data replication
- Configuration version control
- Environment-specific backups

This deployment configuration provides a robust, scalable, and production-ready setup for the Vybes application on Railway's platform.