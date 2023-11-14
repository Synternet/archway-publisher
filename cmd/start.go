package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/syntropynet/archway-publisher/internal/archway"
	"github.com/syntropynet/data-layer-sdk/pkg/service"
)

var (
	flagTendermintAPI *string
	flagRPCAPI        *string
	flagGRPCAPI       *string
)

// startCmd represents the nft command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		name, _ := cmd.Flags().GetString("name")
		publisher := archway.New(
			service.WithContext(ctx),
			service.WithTelemetryPeriod(*flagTelemetryPeriod),
			service.WithName(name),
			service.WithPrefix(*flagPrefixName),
			service.WithNats(natsConnection),
			service.WithUserCreds(*flagUserCreds),
			service.WithNKeySeed(*flagNkey),
			service.WithPemPrivateKey(*flagPemFile),
			service.WithVerbose(*flagVerbose),
			archway.WithTendermintAPI(*flagTendermintAPI),
			archway.WithRPCAPI(*flagRPCAPI),
			archway.WithGRPCAPI(*flagGRPCAPI),
		)

		if publisher == nil {
			return
		}

		pubCtx := publisher.Start()
		defer publisher.Close()

		select {
		case <-ctx.Done():
			log.Println("Shutdown")
		case <-pubCtx.Done():
			log.Println("Publisher stopped with cause: ", context.Cause(pubCtx).Error())
		}
	},
}

func setDefault(field string, value string) {
	if os.Getenv(field) == "" {
		os.Setenv(field, value)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)

	const (
		ARCHWAY_TENDERMINT = "ARCHWAY_TENDERMINT"
		ARCHWAY_RPC        = "ARCHWAY_RPC"
		ARCHWAY_GRPC       = "ARCHWAY_GRPC"
		ARCHWAY_NAME       = "ARCHWAY_SUBJECT"
	)

	setDefault(ARCHWAY_TENDERMINT, "tcp://localhost:26657")
	setDefault(ARCHWAY_RPC, "http://localhost:1317")
	setDefault(ARCHWAY_GRPC, "localhost:9090")
	setDefault(ARCHWAY_NAME, "archway")

	startCmd.Flags().StringP("name", "", os.Getenv(ARCHWAY_NAME), "NATS subject name as in {prefix}.{name}.>")
	flagTendermintAPI = startCmd.Flags().StringP("tendermint-api", "t", os.Getenv(ARCHWAY_TENDERMINT), "Full address to the Tendermint RPC")
	flagRPCAPI = startCmd.Flags().StringP("app-api", "a", os.Getenv(ARCHWAY_RPC), "Full address to the Applications RPC")
	flagGRPCAPI = startCmd.Flags().StringP("grpc-api", "g", os.Getenv(ARCHWAY_GRPC), "Full address to the Applications gRPC")
}
