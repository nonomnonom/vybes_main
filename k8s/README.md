# Vybes Kubernetes Cluster

This directory contains Kubernetes manifests to deploy the Vybes application as a production-ready cluster.

## 🏗️ Architecture

The cluster consists of the following components:

- **Vybes API**: Main application (3 replicas with auto-scaling)
- **MongoDB**: Primary database with persistent storage
- **Redis**: Caching layer
- **MinIO**: Object storage service
- **NATS**: Message broker for async communication
- **Ingress**: External traffic routing
- **HPA**: Horizontal Pod Autoscaler for automatic scaling

## 📋 Prerequisites

1. **Kubernetes Cluster**: Minikube, kind, or any Kubernetes cluster
2. **kubectl**: Kubernetes command-line tool
3. **kustomize**: Kubernetes native configuration management
4. **Docker**: For building the application image
5. **Ingress Controller**: NGINX Ingress Controller (recommended)

## 🚀 Quick Start

### 1. Install Prerequisites

```bash
# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install kustomize
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash
sudo mv kustomize /usr/local/bin/

# Install kind (for local development)
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
sudo install -o root -g root -m 0755 kind /usr/local/bin/kind
```

### 2. Create Local Cluster (Optional)

```bash
# Create kind cluster
kind create cluster --name vybes-cluster

# Verify cluster
kubectl cluster-info
```

### 3. Install Ingress Controller

```bash
# Install NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/baremetal/deploy.yaml

# Wait for ingress controller to be ready
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=120s
```

### 4. Deploy the Cluster

```bash
# Make scripts executable
chmod +x k8s/deploy.sh k8s/cleanup.sh

# Deploy the cluster
cd k8s
./deploy.sh
```

## 📊 Monitoring

### Check Cluster Status

```bash
# View all pods
kubectl get pods -n vybes

# View services
kubectl get services -n vybes

# View deployments
kubectl get deployments -n vybes

# View ingress
kubectl get ingress -n vybes
```

### View Logs

```bash
# View API logs
kubectl logs -f deployment/vybes-api -n vybes

# View MongoDB logs
kubectl logs -f deployment/mongo -n vybes

# View Redis logs
kubectl logs -f deployment/redis -n vybes
```

### Access Services

```bash
# Port forward to API
kubectl port-forward service/vybes-api-service 8080:80 -n vybes

# Port forward to MinIO console
kubectl port-forward service/minio-service 9001:9001 -n vybes

# Port forward to NATS monitoring
kubectl port-forward service/nats-service 8222:8222 -n vybes
```

## 🔧 Configuration

### Environment Variables

The application configuration is managed through ConfigMaps and Secrets:

- **ConfigMap**: `vybes-config` - Non-sensitive configuration
- **Secret**: `vybes-secrets` - Sensitive data (API keys, passwords)

### Scaling

The cluster includes automatic scaling:

- **HPA**: Scales API pods based on CPU (70%) and memory (80%) usage
- **Min Replicas**: 3
- **Max Replicas**: 10

### Storage

- **MongoDB**: 10Gi persistent volume
- **MinIO**: 20Gi persistent volume
- **Storage Class**: `local-storage` (hostPath)

## 🛠️ Customization

### Update Secrets

```bash
# Update secrets with your actual values
kubectl create secret generic vybes-secrets \
  --from-literal=MINIO_ACCESS_KEY=your-access-key \
  --from-literal=MINIO_SECRET_KEY=your-secret-key \
  --from-literal=JWT_SECRET=your-jwt-secret \
  --from-literal=API_KEY=your-api-key \
  -n vybes --dry-run=client -o yaml | kubectl apply -f -
```

### Scale Manually

```bash
# Scale API replicas
kubectl scale deployment vybes-api --replicas=5 -n vybes

# Scale MongoDB replicas
kubectl scale deployment mongo --replicas=3 -n vybes
```

### Update Image

```bash
# Build new image
docker build -t vybes-api:latest .

# Update deployment
kubectl set image deployment/vybes-api vybes-api=vybes-api:latest -n vybes
```

## 🧹 Cleanup

```bash
# Remove entire cluster
cd k8s
./cleanup.sh
```

## 🔍 Troubleshooting

### Common Issues

1. **Image Pull Errors**: Ensure the Docker image is built and available
2. **Storage Issues**: Check if storage directories exist and have proper permissions
3. **Ingress Not Working**: Verify ingress controller is installed and running
4. **Service Connection Issues**: Check if all services are running and healthy

### Debug Commands

```bash
# Describe resources
kubectl describe pod <pod-name> -n vybes
kubectl describe service <service-name> -n vybes

# Check events
kubectl get events -n vybes --sort-by='.lastTimestamp'

# Check resource usage
kubectl top pods -n vybes
kubectl top nodes
```

## 📚 Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Kustomize Documentation](https://kustomize.io/)
- [NGINX Ingress Controller](https://kubernetes.github.io/ingress-nginx/)
- [Horizontal Pod Autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)