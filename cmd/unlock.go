package cmd

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
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
	// prompter.Password("Please Enter Password (Not Yet Implemented, so any string will be accepted!)")
	fmt.Println(yesNo("Warning: Are you sure you would like to unlock your context?"))
	log.Info("Unlocking Context '", kubeContext, "'.")
	setContextStatus(kubeContext, index, "unlocked", config)

	return nil
}

func yesNo(body string) bool {
	prompt := promptui.Select{
		Label: body + " Select[Yes/No]",
		Items: []string{"Yes", "No"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		log.Fatal("Prompt failed %v\n", err)
		os.Exit(1)
	}
	if result != "Yes" {
		log.Fatal("User Answered with 'No', Exiting...")
		os.Exit(1)
	}
	return result == "Yes"
}
