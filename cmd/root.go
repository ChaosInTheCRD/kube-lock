package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile     string
	userLicense string
	context     string

	rootCmd = &cobra.Command{
		Use:   "kube-lock",
		Short: "A pane of glass between you and your Kubernetes clusters.",
		Long:  "kube-lock sits as an intermediary between you and kubectl, allowing you to lock and unlock contexts.\n\nThis aims to prevent misfires to production / high-value Kubernetes clusters that you might have strong IAM privileges on. kube-lock supports custom 'Profiles', allowing you to restrict certain verbs from being passed to high-value clusters. \n\nWARNING: This tool DOES NOT serve as an alternative to Kubernetes Role-Based Access Control, the de-facto standard method of configuring access to the Kubernetes API. This tool provides a convenient layer of protection if you happen to have privileged credentials to a Kubernetes cluster stored locally and an extra layer of protection is preferred.",
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()

}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "verbose logging")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().StringVar(&context, "context", "", "the Kubernetes context you want to address")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".kube-lock" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".kube-lock")

		configFilePath := home + "/.kube-lock.yaml"
		_, err = os.Stat(configFilePath)

		// create file if not exists
		if os.IsNotExist(err) {
			var file, err = os.Create(configFilePath)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			defer file.Close()
			fmt.Println("File Created Successfully", configFilePath)
		}

	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Debug("Using config file:", viper.ConfigFileUsed())
	}
}
