#!/bin/bash

set -e

echo "🚀 Deploying Vybes Kubernetes Cluster..."

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed. Please install kubectl first."
    exit 1
fi

# Check if kustomize is installed
if ! command -v kustomize &> /dev/null; then
    echo "❌ kustomize is not installed. Please install kustomize first."
    exit 1
fi

# Build the Docker image
echo "📦 Building Docker image..."
docker build -t vybes-api:latest .

# Load image into kind cluster (if using kind)
if kubectl config current-context | grep -q "kind"; then
    echo "📥 Loading image into kind cluster..."
    kind load docker-image vybes-api:latest
fi

# Create storage directories
echo "📁 Creating storage directories..."
sudo mkdir -p /data/mongo /data/minio
sudo chmod 777 /data/mongo /data/minio

# Deploy the cluster
echo "🔧 Deploying Kubernetes resources..."
kubectl apply -k .

# Wait for deployments to be ready
echo "⏳ Waiting for deployments to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/mongo -n vybes
kubectl wait --for=condition=available --timeout=300s deployment/redis -n vybes
kubectl wait --for=condition=available --timeout=300s deployment/minio -n vybes
kubectl wait --for=condition=available --timeout=300s deployment/nats -n vybes
kubectl wait --for=condition=available --timeout=300s deployment/vybes-api -n vybes

echo "✅ Vybes cluster deployed successfully!"
echo ""
echo "📊 Cluster Status:"
kubectl get pods -n vybes
echo ""
echo "🌐 Services:"
kubectl get services -n vybes
echo ""
echo "📈 To view logs:"
echo "  kubectl logs -f deployment/vybes-api -n vybes"
echo ""
echo "🔍 To access the API:"
echo "  kubectl port-forward service/vybes-api-service 8080:80 -n vybes"
echo ""
echo "🗑️  To delete the cluster:"
echo "  kubectl delete namespace vybes"