package main

import "github.com/hongfs/prometheus-cloud-target/pkg/cmd"

func main() {
	err := cmd.Start(":9000")

	if err != nil {
		panic(err)
	}
}
