package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ZoneCNH/configx/pkg/configx"
)

type appConfig struct {
	Name    string               `config:"APP_NAME" required:"true"`
	Timeout time.Duration        `config:"TIMEOUT" default:"1s"`
	Token   configx.SecretString `config:"API_TOKEN"`
}

func main() {
	loader := configx.NewLoader().AddSource(configx.NewSecretMapSource("example", map[string]string{
		"APP_NAME":  "configx",
		"API_TOKEN": "fixture-sensitive-value",
	}, []string{"API_TOKEN"}))

	result, err := loader.Load(context.Background())
	if err != nil {
		panic(err)
	}
	var cfg appConfig
	if err := configx.Decode(result, &cfg); err != nil {
		panic(err)
	}

	fmt.Println(cfg.Name)
	fmt.Println(cfg.Timeout)
	fmt.Println(result.Sanitize().Values["API_TOKEN"].Value)
}
