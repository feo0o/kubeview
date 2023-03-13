package cmd

import (
	"github.com/feo0o/kubeview/resource"
	"github.com/spf13/cobra"
)

var resourceCmd = &cobra.Command{
	Use: "resource",
	Run: func(cmd *cobra.Command, args []string) {
		resource.PrintNodesResourcesToStdout()
	},
}
