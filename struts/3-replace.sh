#!/usr/bin/env bash

kubectl apply -f ro-struts.yaml
kubectl get pod -n api
sleep 10
kubectl get pod -n api
