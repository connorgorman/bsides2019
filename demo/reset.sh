#! /bin/bash

kb delete deploy/server ds/listener
kubectl delete deploy/ping deploy/ubuntu
kubectl create -f ../yamls/ping.yaml

sleep 30

kb create -f ../deploy/deploy.yaml
sleep 10
kb create -f ../deploy/listener.yaml

sleep 20
kubectl -n bsides port-forward $(kb get po | grep server | awk '{print $1}') 8080:8080

