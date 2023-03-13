package cmd

import (
    "github.com/feo0o/kubeview/app"
    "github.com/spf13/cobra"
    "strings"
)

var rootCmd = &cobra.Command{
    Use:   strings.ToLower(app.Name),
    Short: "Check status for Kubernetes cluster.",
    Run: func(cmd *cobra.Command, args []string) {
        //
    },
}

func init() {
    rootCmd.AddCommand(
        completionCmd,
        componentCmd,
        healthCmd,
        helpCmd,
        resourceCmd,
        securityCmd,
        versionCmd,
    )
}

func Exec() error {
    return rootCmd.Execute()
}
