package main

import (
	"flag"
	"fmt"
	"github.com/Generalzy/GeneralCache/Cache"
	"github.com/Generalzy/GeneralCache/Cache/cachehttp"
	"log"
	"net/http"
)

func RunServer() {
	var port int
	var hasTom int

	flag.IntVar(&port, "port", 7777, "server port")
	flag.IntVar(&hasTom, "hasTom", 0, "是否含有tom")
	flag.Parse()

	server := cachehttp.NewHTTPServerPool(fmt.Sprintf("127.0.0.1:%d", port))
	server.AddNode("127.0.0.1:7777", "127.0.0.1:8888", "127.0.0.1:9999")

	var c *Cache.CacheGroup
	if hasTom == 1 {
		c = Cache.NewCacheGroup("score", 1<<10, Cache.GetterFunc(func(key string) ([]byte, error) {
			if key == "Tom" {
				return []byte("看你爹做什么"), nil
			}
			return []byte(""), fmt.Errorf("%s not found", key)
		}))
	} else {
		c = Cache.NewCacheGroup("score", 1<<10, Cache.GetterFunc(func(key string) ([]byte, error) {
			return []byte(""), fmt.Errorf("%s not found", key)
		}))
	}
	c.RegisterPickerToCacheGroup(server)
	log.Println(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), server))
}

func main() {
	RunServer()
}
