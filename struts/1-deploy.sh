#!/usr/bin/env bash

echo "Creating Struts-vulnerable deployment..."
kubectl apply -f struts.yaml
kubectl get pod -n api -w
