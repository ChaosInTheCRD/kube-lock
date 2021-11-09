package cmd

import (
	"os"
	"os/exec"
	"strings"
	"time"

	plural "github.com/gertd/go-pluralize"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v3"
	discovery "k8s.io/client-go/discovery/cached/disk"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

func getDeleteBoolFlags() []string {
	return []string{"--all", "--all-namespaces", "--force", "--ignore-not-found", "--now", "--recursive", "-R", "--wait"}
}

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
	Name             string                     `yaml: "name"`
	BlockedVerbs     []string                   `yaml: "blockedVerbs"`
	DeleteExceptions []KubeLockDeleteExceptions `yaml: "deleteExceptions"`
}

type KubeLockDeleteExceptions struct {
	Group    string `yaml: "group"`
	Resource string `yaml: "resource"`
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
			log.Fatal("Context evaluation failed: ", err)
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
		log.Debug("Your context is unlocked! Proceed...", status)
		return true, nil
	} else if status == "locked" {
		log.Error("Halt! Your context is locked! Exiting...")
		os.Exit(1)
	}

	// Checking status has an associated profile
	ok, blockedVerbs, deleteExceptions := validateProfileInConfig(status, config)
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

	// we must check if the verb should be blocked
	if !contains(blockedVerbs, verb) {
		log.Debug("verb '", verb, "' is authorized with Profile ", status, "! Proceed...", status)
		return true, nil
	} else if verb == "delete" {
		log.Debug("Delete exceptions must be checked, continuing...")
	} else {
		log.Error("Halt! Your context has status '", status, "' which is not authorized to '", verb, "' resources! Exiting...")
		os.Exit(1)
	}

	// Finally, we must check if there is a delete exception for the delete command
	kubeconfig := os.Getenv("KUBECONFIG")
	plural := plural.NewClient()
	for _, exception := range deleteExceptions {
		if exception.Resource == resource {
			log.Debug("Delete exception in Profile '", status, "' allows for deleting '", plural.Plural(resource), "'! Proceeding...")
		}
	}
	for _, exception := range deleteExceptions {
		exists, err := findResourceTypeFromDiscovery(kubeconfig, resource, exception.Group, exception.Resource)
		if err != nil {
			return false, err
		}

		if exists {
			log.Debug("Delete exceptions in Profile '", status, "' allows for deleting '", plural.Plural(resource), "'! Proceed...")
			return true, nil
		} else {
			log.Debug("Delete exception '", exception.Resource, "' does not match any resources in group '", exception.Group, "'...")
		}
	}

	log.Error("Halt! Delete exceptions in Profile '", status, "' does not allow for deleting '", plural.Plural(resource), "'! Exiting...")
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

	delBoolFlags := getDeleteBoolFlags()

	for _, arg := range args {
		switch {
		// The next loop must be skipped
		case skipLoop:
			skipLoop = false
			continue
		// A long flag with a space separated value
		case strings.HasPrefix(arg, "--") && !strings.Contains(arg, "="):
			log.Debug(arg, " is a flag with a space separator, skipping the next string and continuing")
			skipLoop = true
			continue
		// A short flag with a space separated value
		case strings.HasPrefix(arg, "-") && !strings.Contains(arg, "=") && len(arg) == 2:
			log.Debug(arg, " is a flag with a space separator, skipping the next string and continuing")
			skipLoop = true
			continue
		case isFlagArg(arg):
			log.Debug(arg, " is a flag with a '=' separator, continuing")
			continue
		// Checking for delete bool flags
		case arg != "" && contains(delBoolFlags, arg):
			log.Debug(arg, "is a delete bool flag, and should be skipped")
			continue
		}

		if verb == "" {
			verb = arg
			if verb != "delete" {
				log.Debug("verb '", verb, "' is not the verb 'delete' and so we are not looking for further rules")
				break
			} else {
				continue
			}
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
		if config.DefaultProfile == "" {
			log.Debug("Ensuring defaults are setup if not already:")
			config.DefaultProfile = "protected"
			config.Profiles = append(config.Profiles, KubeLockProfiles{Name: "protected", BlockedVerbs: []string{"delete", "apply", "create", "patch", "label", "annotate", "replace", "cp", "taint", "drain", "uncordon", "cordon", "auto-scale", "scale", "rollout", "expose", "run", "set"}, DeleteExceptions: []KubeLockDeleteExceptions{{Group: "cert-manager.io/v1", Resource: "certificates"}, {Group: "v1", Resource: "pods"}}})
			WriteToConfig(config)
		}

		log.Warn("kube-lock found that no config entry exists for context '", kubeContext, "'. Adding to config and setting to unlocked and continuing.")
		newContext := KubeLockContexts{Name: kubeContext, Status: "unlocked"}
		config.Contexts = append(config.Contexts, newContext)
		WriteToConfig(config)

	} else if status == "" {
		log.Warn("kube-lock found that context '", kubeContext, "' has no status set, so will set to 'locked' for safety reasons.")
		setContextStatus(kubeContext, contextIndex, "locked", config)
		os.Exit(1)
	}

	return status, contextIndex, nil
}

func findResourceTypeFromDiscovery(kubeConfig string, resource string, groupVersion string, exceptionResource string) (bool, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return false, err
	}

	discoveryClient, err := discovery.NewCachedDiscoveryClientForConfig(config, "/Users/tom/.kube/cache/discovery", "", time.Duration(10*time.Millisecond))
	if err != nil {
		return false, err
	}

	resourceList, err := discoveryClient.ServerResourcesForGroupVersion(groupVersion)

	for _, res := range resourceList.APIResources {
		if res.Name == exceptionResource {
			if (resource != res.Name) || contains(res.ShortNames, resource) || contains(res.Verbs, resource) {
				log.Debug("resource ", resource, " does not match any strings in resource ", res.Name)
			} else {
				log.Debug("resource ", resource, " matches a string in resource ", res.Name)
				return true, nil
			}
		}
	}

	return false, nil
}
