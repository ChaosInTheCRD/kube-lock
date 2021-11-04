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
		Short: "A friendly kubectl wrapper that provides a pain of glass between you and your cluster.",
		Long:  `Still need to write...`,
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

	// rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	// viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	// viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	// viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	// viper.SetDefault("license", "apache")

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
