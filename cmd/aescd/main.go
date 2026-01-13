package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/extend8888/aescd/cmd/aescd/cmd"
	"github.com/extend8888/aescd/cmd/aescd/config"
)

func main() {
	setupSDKConfig()

	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "aescd", config.MustGetDefaultNodeHome()); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}

func setupSDKConfig() {
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)
	cfg.Seal()
}
