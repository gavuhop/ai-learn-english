package main

import (
	"fmt"

	"ai-learn-english/config"
)

func main() {
	config.Init("config.yaml")
	fmt.Println("Hello, World!")
}