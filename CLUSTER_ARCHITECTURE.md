# Vybes API - Cluster Architecture

## 🏗️ **High Availability Cluster Setup**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              LOAD BALANCER                                  │
│                              (Railway Auto)                                 │
└─────────────────────┬───────────────────────────────────────────────────────┘
                      │
        ┌─────────────┴─────────────┐
        │                           │
┌───────▼────────┐        ┌────────▼────────┐
│    API-1       │        │     API-2       │
│  (Go/Gin)      │        │   (Go/Gin)      │
│  Port: 8080    │        │   Port: 8080    │
│  [Auto-scale]  │        │  [Auto-scale]   │
└───────┬────────┘        └────────┬────────┘
        │                          │
        └──────────┬───────────────┘
                   │
    ┌──────────────┴──────────────┐
    │      WORKER CLUSTER         │
    │   (MongoDB + Redis + NATS)  │
    └──────────────┬──────────────┘
                   │
    ┌──────────────┼──────────────┐
    │              │              │
┌───▼───┐    ┌────▼────┐    ┌────▼────┐
│WORKER-1│    │WORKER-2 │    │WORKER-3 │
│MongoDB │    │MongoDB  │    │MongoDB  │
│Redis   │    │Redis    │    │Redis    │
│NATS    │    │NATS     │    │NATS     │
│Ports:  │    │Ports:   │    │Ports:   │
│27017   │    │27017    │    │27017    │
│6379    │    │6379     │    │6379     │
│4222    │    │4222     │    │4222     │
└───────┘    └─────────┘    └─────────┘
    │              │              │
    └──────────────┼──────────────┘
                   │
    ┌──────────────┴──────────────┐
    │      EXTERNAL STORAGE       │
    │      (Cloudflare R2)        │
    └─────────────────────────────┘
```

## 📊 **Service Breakdown**

### **API Layer (2 Nodes)**
| Service | Type | Port | Scaling | Purpose |
|---------|------|------|---------|---------|
| `api-1` | Go App | 8080 | Auto-scale | Primary API |
| `api-2` | Go App | 8080 | Auto-scale | Backup API |

### **Worker Layer (3 Nodes)**
| Service | MongoDB | Redis | NATS | Purpose |
|---------|---------|-------|------|---------|
| `worker-1` | ✅ | ✅ | ✅ | Primary worker |
| `worker-2` | ✅ | ✅ | ✅ | Secondary worker |
| `worker-3` | ✅ | ✅ | ✅ | Tertiary worker |

### **External Storage (Cloudflare R2)**
| Service | Type | Replication | Purpose |
|---------|------|-------------|---------|
| Cloudflare R2 | S3-Compatible | Global CDN | Object storage for posts, stories, and media |

## 🔄 **High Availability Features**

### **1. Load Balancing**
- **Railway Auto-LB**: Automatically distributes traffic between API nodes
- **Health Checks**: Failed nodes are removed from rotation
- **Auto-scaling**: Additional API instances created based on demand

### **2. Database Replication**
- **MongoDB Replica Set**: 3-node replication for data safety
- **Automatic Failover**: If primary fails, secondary takes over
- **Read Distribution**: Reads can be distributed across nodes

### **3. Cache Distribution**
- **Redis Cluster**: 3-node Redis cluster for high availability
- **Data Sharding**: Data distributed across nodes
- **Failover**: Automatic failover if node goes down

### **4. Message Queue Redundancy**
- **NATS Cluster**: 3-node NATS cluster for message reliability
- **Message Persistence**: Messages stored across multiple nodes
- **Auto-recovery**: Failed nodes automatically rejoin cluster

### **5. External Storage (R2)**
- **Cloudflare R2**: S3-compatible object storage
- **Global CDN**: Content delivered from edge locations worldwide
- **High Durability**: 99.999999999% (11 9's) durability
- **Cost Effective**: No egress fees, pay only for storage and requests

## 🚀 **Performance Benefits**

### **Scalability**
- **Horizontal Scaling**: Add more nodes as needed
- **Load Distribution**: Traffic spread across multiple nodes
- **Resource Optimization**: Each node handles specific workload

### **Reliability**
- **99.9% Uptime**: Multiple nodes ensure service availability
- **Zero Downtime**: Rolling updates without service interruption
- **Data Safety**: Multiple copies of data across nodes

### **Performance**
- **Reduced Latency**: Closer nodes to users
- **Parallel Processing**: Multiple nodes handle requests simultaneously
- **Caching**: Distributed cache reduces database load
- **Global CDN**: R2 provides fast content delivery worldwide

## 🔧 **Configuration Details**

### **MongoDB Replica Set**
```bash
# Connection string
mongodb://worker-1:27017,worker-2:27017,worker-3:27017/vybes_production?replicaSet=vybes-rs

# Replica set configuration
- Primary: worker-1
- Secondary: worker-2, worker-3
- Automatic failover enabled
```

### **Redis Cluster**
```bash
# Connection string
worker-1:6379,worker-2:6379,worker-3:6379

# Cluster configuration
- 3 master nodes
- Automatic sharding
- Failover enabled
```

### **Cloudflare R2**
```bash
# Endpoint
https://{ACCOUNT_ID}.r2.cloudflarestorage.com

# Configuration
- S3-compatible API
- Global CDN integration
- No egress fees
- Automatic scaling
```

### **NATS Cluster**
```bash
# Connection string
nats://worker-1:4222,nats://worker-2:4222,nats://worker-3:4222

# Cluster configuration
- 3-node cluster
- Message persistence
- Automatic node discovery
```

## 💰 **Cost Optimization**

### **Resource Allocation**
- **Pay-per-use**: Only pay for actual resource consumption
- **Auto-scaling**: Resources scale with demand
- **Efficient**: No over-provisioning

### **Estimated Monthly Cost**
- **API Nodes**: $20-200 (depending on traffic)
- **Worker Nodes**: $30-150 (database + cache + queue)
- **R2 Storage**: $5-50 (depending on storage usage, no egress fees)
- **Total**: $55-400/month for enterprise-grade setup

## 🛡️ **Security Features**

### **Network Security**
- **Internal Networking**: All inter-node communication encrypted
- **Service Isolation**: Each service runs in isolated environment
- **Access Control**: Role-based access to different services

### **Data Security**
- **Encryption at Rest**: All data encrypted on disk
- **Encryption in Transit**: TLS for all communications
- **Authentication**: Strong authentication for all services
- **R2 Security**: IAM policies, bucket policies, and encryption

This cluster architecture provides enterprise-grade reliability, scalability, and performance for the Vybes application with cost-effective external storage!