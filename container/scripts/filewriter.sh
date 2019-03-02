#! /bin/bash

mkdir /var/lib/data

for n in {1..1000}; do
    dd if=/dev/urandom of=/var/lib/data/file$( printf %03d "$n" ).bin bs=1 count=$(( RANDOM + 1024 ))
    sleep 5
done
