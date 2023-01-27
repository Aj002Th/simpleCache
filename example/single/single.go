package main

import (
	"fmt"
	"log"
	"net/http"
	"simpleCache"
)

var dbSingle = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	simpleCache.NewGroup("scores", 2<<10, simpleCache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := dbSingle[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := simpleCache.NewHttpPool(addr)
	log.Println("simpelcache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
