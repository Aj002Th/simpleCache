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
	sim := simpleCache.NewGroup("scores", 2<<10, simpleCache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := dbSingle[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "http://localhost:8001"
	apiAddr := "http://localhost:9999"
	peers := simpleCache.NewHttpPool(addr)
	peers.Set(addr)
	sim.RegisterPeerPicker(peers)

	// api server
	go func() {
		http.Handle("/api", http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				key := r.URL.Query().Get("key")
				view, err := sim.Get(key)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/octet-stream")
				_, _ = w.Write(view.ByteSlice())
			}))
		log.Println("fontend server is running at", apiAddr)
		log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
	}()

	// peer server
	log.Println("simpelcache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}
