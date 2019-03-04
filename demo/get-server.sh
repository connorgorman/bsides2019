#! /bin/bash

curl -sk http://localhost:8081/containers/server | jq .
