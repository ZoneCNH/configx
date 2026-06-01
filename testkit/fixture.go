package testkit

import (
	"time"

	"github.com/ZoneCNH/configx/pkg/configx"
)

func Config(name string) configx.Config {
	return configx.Config{
		Name:    name,
		Timeout: time.Second,
	}
}
