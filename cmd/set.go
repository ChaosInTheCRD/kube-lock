package cmd

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setCmd)
}

var setCmd = &cobra.Command{
	Use:    "set",
	Short:  "Set a profile from the kube-lock config as the status for a context.",
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		nativeCmd = true
		err := setProfile(cmd, args)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func setProfile(cmd *cobra.Command, args []string) error {
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

	ok, blockedVerbs, deleteExceptions := validateProfileInConfig(args[0], config)
	if !ok {
		log.Error("Profile '", args[0], "' not found. Please add it, or change Profile for context '", kubeContext, "'.")
		os.Exit(1)
	}

	log.Info("Setting Status '", args[0], "' for context '", kubeContext, "'.")
	setContextStatus(kubeContext, index, args[0], config)

	blockedVerbsOut := "'" + strings.Join(blockedVerbs, `','`) + `'`
	log.Info("\nProfile Rules:")
	log.Info("Blocked Verbs: ", blockedVerbsOut)
	log.Info("Delete Exceptions: ", deleteExceptions)
	return nil
}

func validateProfileInConfig(profile string, config KubeLockConfig) (bool, []string, []KubeLockDeleteExceptions) {
	log.Debug("Validating that Profile '", profile, "' exists in kube-lock config.")
	var blockedVerbs []string
	var deleteExceptions []KubeLockDeleteExceptions
	var ok bool
	for i, profiles := range config.Profiles {
		if profiles.Name == profile {
			blockedVerbs = config.Profiles[i].BlockedVerbs
			deleteExceptions = config.Profiles[i].DeleteExceptions
			ok = true
		}
	}

	return ok, blockedVerbs, deleteExceptions
}
