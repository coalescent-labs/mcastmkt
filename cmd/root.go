package cmd

import (
	"fmt"
	"github.com/coalescent-labs/mcastmkt/cmd/any"
	"github.com/coalescent-labs/mcastmkt/cmd/eurex"
	"github.com/coalescent-labs/mcastmkt/cmd/euronext"
	"github.com/coalescent-labs/mcastmkt/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	cfgFile string

	mcastmktCmd = &cobra.Command{
		Use:           "mcastmkt",
		Short:         "mcastmkt – command-line tool to test markets multicast traffic flows",
		Long:          ``,
		Version:       version.String(),
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func Execute() error {
	return mcastmktCmd.Execute()
}

func init() {
	// Allow to load some default config from file (unused for now)
	cobra.OnInitialize(initConfig)
	mcastmktCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.mcastmkt.yaml)")
	_ = viper.BindPFlag("config", mcastmktCmd.PersistentFlags().Lookup("config"))

	// Add subcommands here
	mcastmktCmd.AddCommand(any.AnyCmd)
	mcastmktCmd.AddCommand(eurex.EurexCmd)
	mcastmktCmd.AddCommand(euronext.EuronextCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigFile(".mcastmkt")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Config file used for mcastmkt: ", viper.ConfigFileUsed())
	}
}
