package cmd

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setTimeoutCmd)
}

var setTimeoutCmd = &cobra.Command{
	Use:    "set-timeout",
	Short:  "An easy way to set the unlock timeout duration (e.g. '10s', '10m', '10h' etc.).",
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		nativeCmd = true
		err := setTimeout(cmd, args)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func setTimeout(cmd *cobra.Command, args []string) error {
	var err error
	newTimeout := args[0]

	config, err := getViperConfig()
	if err != nil {
		return err
	}

	log.Info("Setting new Unlock Timeout Period to '", newTimeout, "'...")
	_, err = time.ParseDuration(newTimeout)
	if err != nil {
		log.Error("Provided Timeout Value '", newTimeout, "' not provided correctly... Exiting")
		return err
	}

	config.UnlockTimeoutPeriod = newTimeout
	err = WriteToConfig(config)
	if err != nil {
		return err
	}

	return nil
}
