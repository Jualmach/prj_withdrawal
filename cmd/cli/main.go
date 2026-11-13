package main

import (
	"flag"
	"fmt"
	"os"

	"prj_withdrawal/internal/bank"
	"prj_withdrawal/internal/cli"
	"prj_withdrawal/internal/storage"
)

func main() {
	dataPath := flag.String("data", "data/state.json", "ruta del archivo JSON de cuenta y transacciones")
	flag.Parse()

	service := bank.NewService(storage.NewJSONStore(*dataPath))
	if err := cli.Run(os.Stdin, os.Stdout, service, bank.DefaultAccountID); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
