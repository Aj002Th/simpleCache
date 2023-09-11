#!/bin/bash
trap "rm server;kill 0" EXIT

go build -o server
./server &

sleep 2
echo ">>> start test"
curl "http://localhost:9999/api?key=Tom" #630
curl "http://localhost:9999/api?key=kkk" #not exist

# old test for peer
#curl http://localhost:9999/_simplecache/scores/Tom
#curl http://localhost:9999/_simplecache/scores/kkk

wait