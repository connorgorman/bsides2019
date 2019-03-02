#! /bin/bash

echo "TIME      UID    PID    TID    COMM             CAP  NAME                 AUDIT"
for i in {1..1000}; do
    echo "20:10:37  0      26214  26214  python           21   CAP_SYS_ADMIN        1"
    echo "20:10:41  0      26245  26245  ipset            12   CAP_NET_ADMIN        1"
    sleep 5
done
