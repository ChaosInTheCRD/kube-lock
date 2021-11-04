package cmd

import (
	"github.com/Songmu/prompter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(unlockCmd)
}

var unlockCmd = &cobra.Command{
	Use:    "unlock",
	Short:  "An easy way to set the status for a context to 'unlocked'.",
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		nativeCmd = true
		removeLock(cmd, args)
	},
}

func removeLock(cmd *cobra.Command, args []string) error {
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

	// Need to make this actually work
	prompter.Password("Please Enter Password")
	log.Info("Unlocking Context '", kubeContext, "'.")
	setContextStatus(kubeContext, index, "unlocked", config)

	return nil
}
