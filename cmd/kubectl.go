package cmd

import (
	"os"
	"os/exec"
	"strings"

	plural "github.com/gertd/go-pluralize"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v3"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeLockConfig struct {
	Contexts       []KubeLockContexts `yaml: "contexts"`
	Profiles       []KubeLockProfiles `yaml: "profiles"`
	DefaultProfile string             `yaml: "defaultProfile"`
}

type KubeLockContexts struct {
	Name            string `yaml: "name"`
	Status          string `yaml: "status"`
	UnlockTimestamp string `yaml: "unlockTimestamp"`
}

type KubeLockProfiles struct {
	Name         string   `yaml: "name"`
	BlockedVerbs []string `yaml: "blockedVerbs"`
	DeletePods   bool     `yaml: "deletePods"`
}

func init() {
	rootCmd.AddCommand(kubectlCmd)
}

var kubectlCmd = &cobra.Command{
	Use:    "kubectl",
	Short:  "The kubectl command you want to issue when using kube-lock",
	PreRun: toggleDebug,
	Run: func(cmd *cobra.Command, args []string) {
		ok, err := evaluateContext(cmd, args)
		if err != nil {
			log.Fatal("Context evaluation failed: %s", err)
			os.Exit(1)
		}
		if ok {
			execKubectl(cmd, args)
		}
	},
}

func findContextArg(args []string) string {
	var context string
	for i := range args {
		if args[i] == "--context" {
			context = args[i+1]
		}
	}

	return context
}

func findContextConfig() (string, error) {
	configPath := os.Getenv("KUBECONFIG")
	kubeConfig, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return "", err
	}
	return kubeConfig.CurrentContext, nil
}

func findContext(args []string) (string, error) {
	// First we want to evaluate if the user has specified a context
	kubeContext := findContextArg(args)
	if kubeContext != "" {
		return kubeContext, nil
	} else {
		var err error
		kubeContext, err = findContextConfig()
		if err != nil {
			return kubeContext, err
		}
	}

	return kubeContext, nil
}

func evaluateContext(cmd *cobra.Command, args []string) (bool, error) {
	plural := plural.NewClient()

	// Finding the current context set
	kubeContext, err := findContext(args)
	if err != nil {
		return false, err
	}

	// Getting the kube-lock config from viper
	config, err := getViperConfig()
	if err != nil {
		return false, err
	}

	status, _, err := findContextInConfig(kubeContext, config)
	if err != nil {
		return false, err
	}

	// Exit now if status is 'unlocked' or 'locked'
	if status == "unlocked" {
		log.Debug("Your context is unlocked, proceed", status)
		return true, nil
	} else if status == "locked" {
		log.Warn("Halt! Your context is locked! Exiting...")
		os.Exit(1)
	}

	// Checking status has an associated profile
	ok, blockedVerbs, deletePods := validateProfileInConfig(status, config)
	if ok != true {
		log.Error("Profile '", status, "' not found. Please add it, or change Profile for context '", kubeContext, "'.")
		os.Exit(1)
	}

	// Find the verb and resource strings from the kubectl command issued by the user
	var verb string
	var resource string
	verb, resource, err = findArgs(args)
	if err != nil {
		return false, err
	}

	// If there is a 'delete pod' command issued of some sort, check if the 'deletePods' boolean has been set
	if (verb == "delete" && contains([]string{"pod", "pods", "po"}, resource)) || strings.Contains(resource, "pod/") || strings.Contains(resource, "pods/") {
		if deletePods == true {
			log.Debug("You are authorized to delete pods under status %s, proceed", status)
			return true, nil
		} else {
			log.Info("Halt! Your context has status %s, which is not authorized to delete pods! Exiting...", status)
			os.Exit(1)
		}
	}

	// Finally, we must check if the verb should be blocked
	if !contains(blockedVerbs, verb) {
		log.Debug("verb %s is authorized under status %s, proceed", verb, status)
		return true, nil
	} else {
		log.Info("Halt! Your context has status '", status, "' which is not authorized to ", verb, " ", plural.Plural(resource), "! Exiting...")
		os.Exit(1)
	}

	return false, nil
}

// Execute the kubectl command
func execKubectl(cmd *cobra.Command, args []string) {

	kubectlCmd := exec.Command("kubectl", args...)
	kubectlCmd.Stdin = os.Stdin
	kubectlCmd.Stdout = os.Stdout
	kubectlCmd.Stderr = os.Stderr

	kubectlCmd.Start()
	err := kubectlCmd.Wait()

	if err != nil {
		os.Exit(err.(*exec.ExitError).ExitCode())
	}
}

// contains checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// Manipulated from https://github.com/spf13/cobra/blob/bfacc59f62c67ffd43e93655a8d933cefab0fa99/command.go#L685 to find the flags and skip them
func findArgs(args []string) (string, string, error) {
	skipLoop := false
	var verb string
	var resource string

	for _, arg := range args {
		switch {
		// The next loop must be skipped
		case skipLoop:
			skipLoop = false
			continue
		// A long flag with a space separated value
		case strings.HasPrefix(arg, "--") && !strings.Contains(arg, "="):
			log.Warn("%s is a flag with a space separator, skipping the next string and continuing", arg)
			skipLoop = true
			continue
		// A short flag with a space separated value
		case strings.HasPrefix(arg, "-") && !strings.Contains(arg, "=") && len(arg) == 2:
			log.Warn("%s is a flag with a space separator, skipping the next string and continuing", arg)
			skipLoop = true
			continue
		case isFlagArg(arg):
			log.Warn("%s is a flag with a '=' separator, continuing", arg)
			continue
		}

		if verb == "" {
			verb = arg
			continue
		} else if resource == "" {
			resource = arg
			break
		}
	}
	return verb, resource, nil
}

func isFlagArg(arg string) bool {
	return ((len(arg) >= 3 && arg[1] == '-') ||
		(len(arg) >= 2 && arg[0] == '-' && arg[1] != '-'))
}

// There might be a good way of doing this with viper, but this will do for now
func WriteToConfig(config KubeLockConfig) error {
	newConfig, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}
	err = os.WriteFile(viper.ConfigFileUsed(), newConfig, 0)
	if err != nil {
		return err
	}

	return nil
}

func getViperConfig() (KubeLockConfig, error) {
	config := KubeLockConfig{}
	err := viper.Unmarshal(&config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func findContextInConfig(kubeContext string, config KubeLockConfig) (string, int, error) {
	// Getting the lock status for current context
	var status string
	var contextIndex int
	var found bool
	for i, context := range config.Contexts {
		if context.Name == kubeContext {
			status = config.Contexts[i].Status
			found = true
			contextIndex = i
			break
		}
	}

	// If the status isn't populated, add the context to the config with defaults if it doesn't exist
	// If it does exist, but there is no status field populated, lock it to be safe
	if found == false {
		defaultProfile := config.DefaultProfile
		log.Warn("kube-lock found that context '", kubeContext, "' has no config. Loading default profile '", defaultProfile, "'.")
		newContext := KubeLockContexts{Name: kubeContext, Status: defaultProfile}
		config.Contexts = append(config.Contexts, newContext)
		WriteToConfig(config)
		os.Exit(1)
	} else if status == "" {
		log.Warn("kube-lock found that context '", kubeContext, "' has no status set, so will set to 'locked' for safety reasons.")
		setContextStatus(kubeContext, contextIndex, "locked", config)
		os.Exit(1)
	}

	return status, contextIndex, nil
}
