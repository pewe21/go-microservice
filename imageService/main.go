package main

import "sync"

type AppImage struct {
	Filename string `json:"filename"`
}

func main() {

	var wg sync.WaitGroup

	cfg := InitConfig()

	httpServer := NewAppServer(cfg)

	wg.Add(1)
	go func() {
		defer wg.Done()
		httpServer.Run()
	}()

	wg.Wait()
}
