#!/bin/bash

set -e

echo "🧹 Cleaning up Vybes Kubernetes Cluster..."

# Delete all resources in the vybes namespace
echo "🗑️  Deleting all resources..."
kubectl delete namespace vybes --ignore-not-found=true

# Wait for namespace deletion
echo "⏳ Waiting for namespace deletion..."
kubectl wait --for=delete namespace/vybes --timeout=300s || true

# Remove storage directories (optional)
read -p "Do you want to remove storage directories? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "📁 Removing storage directories..."
    sudo rm -rf /data/mongo /data/minio
fi

echo "✅ Cleanup completed!"