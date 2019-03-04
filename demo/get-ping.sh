#! /bin/bash

curl -sk http://localhost:8080/containers/ping | jq .
