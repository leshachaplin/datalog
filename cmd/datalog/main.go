package main

import (
	"github.com/leshachaplin/datalog/app"
	"github.com/leshachaplin/datalog/internal/config"
)

func main() {
	app.New(func() (config.Config, error) {
		return config.Config{}, nil
	}).Start()
}
