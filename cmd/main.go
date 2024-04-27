package main

import (
	"tribler-arr-shim/cmd/server"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tribler-arr-shim",
	Short: "tribler-arr-shim",
	Run: func(cmd *cobra.Command, args []string) {
		server.StartServer()
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run tribler-arr-shim as a server",
	Run: func(cmd *cobra.Command, args []string) {
		server.StartServer()
	},
}

func Execute() {
	rootCmd.AddCommand(serverCmd)
	rootCmd.Execute()
}

func main() {
	Execute()
}
