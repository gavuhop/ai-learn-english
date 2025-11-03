package main

import (
	"ai-learn-english/config"
	"fmt"
)

func main() {
	config.Init("config.yml")
	fmt.Println("Database Host: ", config.Cfg.Database.Host)

}
