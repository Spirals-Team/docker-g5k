package main

import (
	"fmt"
	"os"

	"github.com/Spirals-Team/docker-g5k/command"
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println("	create-cluster <g5k-user> <g5k-password> <g5k-site> <g5k-walltime> <path to ssh private key> <nb nodes>")
	os.Exit(1)
}

func main() {
	cmd := &command.Command{}
	if err := cmd.ParseArguments(os.Args[2:]); err != nil {
		panic(err)
	}

	switch os.Args[1] {
	case "create-cluster":
		cmd.CreateCluster()
	default:
		usage()
	}
}
