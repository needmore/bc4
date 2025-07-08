package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "bc4",
	Short: "A CLI tool for interacting with Basecamp 4 API",
	Long: `bc4 is a command-line interface for Basecamp 4 that allows you to:
- List and manage projects
- Create and manage todos
- Post messages and comments
- Manage campfires and check-ins
- And much more!`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bc4.yaml)")
	rootCmd.PersistentFlags().String("account-id", "", "Basecamp account ID")
	rootCmd.PersistentFlags().String("access-token", "", "Basecamp access token")
	
	viper.BindPFlag("account_id", rootCmd.PersistentFlags().Lookup("account-id"))
	viper.BindPFlag("access_token", rootCmd.PersistentFlags().Lookup("access-token"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".bc4")
	}

	viper.SetEnvPrefix("BC4")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}