package cmd

import (
	"github.com/seatgeek/aws-dynamic-consul-catalog/service/rds"
	"time"

	"github.com/spf13/cobra"
)

var rdsCmd = &cobra.Command{
	Use:   "rds",
	Short: "Dynamically sync AWS RDS information into a Consul Service Catalog",
	Run: func(cmd *cobra.Command, args []string) {
		app := rds.New(cmd)
		app.Run()
	},
}

func init() {
	rootCmd.AddCommand(rdsCmd)
	rdsCmd.Flags().String("consul-leader-tag", "leader", "The Consul service tag for leader instances")
	rdsCmd.Flags().String("consul-follower-tag", "follower", "The Consul service tag for follower instances")
	rdsCmd.Flags().String("consul-node-name", "rds", "Consul catalog node name")
	rdsCmd.Flags().Duration("rds-tag-cache-time", 30*time.Minute, "The time RDS tags should be cached (eg. 30s, 1h, 1h10m, 1d)")
}
