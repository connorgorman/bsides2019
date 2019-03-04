#! /bin/bash

kubectl exec -it $(kubectl get po | grep ping | awk '{print $1}') /bin/bash


