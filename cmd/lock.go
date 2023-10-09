package cmd

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(lockCmd)
}

var lockCmd = &cobra.Command{
	Use:    "lock",
	Short:  "An easy way to set the status for a context to 'locked'.",
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		nativeCmd = true
		err := setLock(cmd, args)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func setLock(cmd *cobra.Command, args []string) error {
	var kubeContext string
	var err error

	if context != "" {
		kubeContext = context
	} else {
		kubeContext, err = findContext(args)
	}
	if err != nil {
		return err
	}

	config, err := getViperConfig()
	if err != nil {
		return err
	}

	_, _, index, err := findContextInConfig(kubeContext, config)
	if err != nil {
		return err
	}

	log.Info("Locking Context '", kubeContext, "'.")
	setContextStatus(kubeContext, index, "locked", config)

	return nil
}

func setContextStatus(kubeContext string, index int, status string, config KubeLockConfig) {
	if status == ("unlocked") && config.Contexts[index].Status == "locked" {
		log.Debug("Setting context from 'locked' to 'unlocked', marking unlock Timestamp to check for timeout later...")
		config.Contexts[index].UnlockTimestamp = time.Now().Format(timestampLayout)
	} else if config.Contexts[index].UnlockTimestamp != "" {
		log.Debug("Clearing unlock timestamp...")
		config.Contexts[index].UnlockTimestamp = ""
	}
	config.Contexts[index].Status = status
	err := WriteToConfig(config)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Info("Set context '", kubeContext, "' to ", status, ".")
}
