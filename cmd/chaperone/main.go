package main

import (
	"context"

	"github.com/KillianMeersman/chaperone/internal/chaperone"
	"github.com/KillianMeersman/chaperone/pkg/telemetry/log"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "",
		Short: "",
	}

	proxyCmd = &cobra.Command{
		Use:   "proxy",
		Short: "Start & serve the Chaperone proxy",
		Long: `
		Starts the proxy and listens on $PORT (default 8080).
		`,
		Run: func(cmd *cobra.Command, args []string) {
			proxy := &chaperone.ChaperoneProxy{}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := proxy.Start(ctx)
			if err != nil {
				log.DefaultLogger.Fatal(ctx, err)
			}
		},
	}
)

func main() {
	rootCmd.AddCommand(proxyCmd)
	ctx := context.Background()
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		log.Fatal(ctx, err)
	}
}
