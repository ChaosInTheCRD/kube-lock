package cmd

import (
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
		setLock(cmd, args)
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

	_, index, err := findContextInConfig(kubeContext, config)
	if err != nil {
		return err
	}

	log.Info("Locking Context '", kubeContext, "'.")
	setContextStatus(kubeContext, index, "locked", config)

	return nil
}

func setContextStatus(kubeContext string, index int, status string, config KubeLockConfig) {
	config.Contexts[index].Status = status
	WriteToConfig(config)
	log.Info("Set context '", kubeContext, "' to ", status, ".")
}
