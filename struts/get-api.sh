#! /bin/bash

curl -sk http://localhost:8081/containers/api | jq .
