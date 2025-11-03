package main

import (
	"ai-learn-english/config"
	"context"
	"fmt"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
)

func connectMilvusWithRetry(address string, attempts int, perAttemptTimeout time.Duration, delay time.Duration) (client.Client, error) {
	var lastErr error
	for i := 0; i < attempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), perAttemptTimeout)
		cli, err := client.NewClient(ctx, client.Config{Address: address})
		cancel()
		if err == nil {
			return cli, nil
		}
		lastErr = err
		time.Sleep(delay)
	}
	return nil, lastErr
}

func main() {
	config.Init("config.yml")
	fmt.Println("Database Host: ", config.Cfg.Database.Host)

	// Milvus demo with retry (Milvus may take tens of seconds to boot)
	cli, err := connectMilvusWithRetry("localhost:19530", 20, 5*time.Second, 2*time.Second)
	if err != nil {
		fmt.Println("Milvus connect error:", err)
		return
	}
	defer cli.Close()

	fmt.Println("Milvus connected!")

}
