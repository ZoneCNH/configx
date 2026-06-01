package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ZoneCNH/configx/pkg/configx"
)

func main() {
	client, err := configx.New(context.Background(), configx.Config{Name: "configx"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "create client: %v\n", err)
		return
	}

	status := client.HealthCheck(context.Background())
	fmt.Println(status.Status)
}
