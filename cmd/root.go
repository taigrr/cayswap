/*
Copyright © 2022 Tai Groot
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/taigrr/cayswap/util"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "cayswap",
	Short: "exchange WireGuard keys automatically and painlessly",
	Long: `cayswap facilitates easy, automated swapping of wireguard public keys
	between authenticated nodes using a preshared key which is only valid for a
	few minutes at a time. Designed to be used in concert with Terraform.`,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	if !util.IsRoot() {
		return
	}

	rootCmd.PersistentFlags().StringP("auth", "k", "", "Auth Key")
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/cayswap/cayswap.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("/etc/cayswap/")
		viper.SetConfigType("yaml")
		viper.SetConfigName("cayswap")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
