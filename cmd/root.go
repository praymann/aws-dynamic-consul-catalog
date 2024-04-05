package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aws-dynamic-consul-catalog",
	Short: "Easily maintain AWS information in a Consul Service Catalog",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringSlice("instance-filter", []string{}, "AWS Instance filter")
	rootCmd.PersistentFlags().StringSlice("tag-filter", []string{}, "AWS Tag filter")
	rootCmd.PersistentFlags().String("consul-service-prefix", "", "Service prefix within Consul Catalog")
	rootCmd.PersistentFlags().String("consul-service-suffix", "", "Service suffix within Consul Catalog")
	rootCmd.PersistentFlags().String("on-duplicate", "ignore-skip-last", "What to do if duplicate services/check are found in RDS (e.g. multiple instances with same DB name or consul_service_name tag - and same RDS Replication Role")
	rootCmd.PersistentFlags().Duration("check-interval", 60*time.Second, "How often to check for changes (eg. 30s, 1h, 1h10m, 1d)")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level")
	rootCmd.PersistentFlags().String("log-format", "text", "Log format")
}
