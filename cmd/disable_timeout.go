package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(disableTimeoutCmd)
}

var disableTimeoutCmd = &cobra.Command{
	Use:    "disable-timeout",
	Short:  "An easy way to set the disable the unlock timeout feature.",
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		nativeCmd = true
		err := disableTimeout()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func disableTimeout() error {

	config, err := getViperConfig()
	if err != nil {
		return err
	}

	log.Info("Disabling Unlock Timeouts...")
	config.UnlockTimeoutPeriod = ""
	err = WriteToConfig(config)
	if err != nil {
		return err
	}

	log.Info("Removing Unlock Timestamps for all contexts...")
	for i := range config.Contexts {
		config.Contexts[i].UnlockTimestamp = ""
	}

	err = WriteToConfig(config)
	if err != nil {
		return err
	}

	return nil
}
