package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bytechainx/configx/pkg/configx"
)

func main() {
	client, err := configx.New(context.Background(), configx.Config{Name: "configx"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "create client: %v\n", err)
		return
	}
	defer func() {
		_ = client.Close(context.Background())
	}()

	fmt.Println(configx.ModuleName)
}
