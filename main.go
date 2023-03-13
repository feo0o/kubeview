package main

import (
	"fmt"
	"os"

	"github.com/feo0o/kubeview/cmd"
	_ "github.com/feo0o/kubeview/kube"
)

func main() {
	if err := cmd.Exec(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
