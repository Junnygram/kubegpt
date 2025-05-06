package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	kubeconfig  string
	namespace   string
	outputFormat string
	slackWebhook string
	reportFile  string
	verbose     bool
	fix         bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubegpt",
	Short: "AI-powered Kubernetes troubleshooting assistant",
	Long: `KubeGPT is an AI-powered Kubernetes troubleshooting assistant that helps
DevOps engineers and SREs diagnose and fix issues in their Kubernetes clusters.

It connects to your Kubernetes cluster, identifies unhealthy resources,
and uses Amazon Q Developer to explain errors and suggest fixes.

Examples:
  # Diagnose issues in the current namespace
  kubegpt diagnose

  # Diagnose issues in a specific namespace
  kubegpt diagnose --namespace monitoring

  # Explain a specific error
  kubegpt explain "CrashLoopBackOff: container exited with code 1"

  # Generate a report of all issues
  kubegpt report --output markdown --file cluster-health.md
`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, print help
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kubegpt.yaml)")
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "kubernetes namespace to use")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "terminal", "output format (terminal, markdown, slack)")
	rootCmd.PersistentFlags().StringVar(&slackWebhook, "slack-webhook", "", "slack webhook URL for notifications")
	rootCmd.PersistentFlags().StringVarP(&reportFile, "file", "f", "", "file to write the report to")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&fix, "fix", false, "generate YAML patches to fix issues")

	// Bind flags to viper
	viper.BindPFlag("kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))
	viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("slack-webhook", rootCmd.PersistentFlags().Lookup("slack-webhook"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("fix", rootCmd.PersistentFlags().Lookup("fix"))

	// Set default kubeconfig path
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			defaultKubeconfig := filepath.Join(home, ".kube", "config")
			if _, err := os.Stat(defaultKubeconfig); err == nil {
				viper.SetDefault("kubeconfig", defaultKubeconfig)
			}
		}
	}

	// Add commands
	rootCmd.AddCommand(diagnoseCmd)
	rootCmd.AddCommand(explainCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(versionCmd)
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".kubegpt" (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".kubegpt")
	}

	// Read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}
}

// printLogo prints the KubeGPT ASCII logo
func printLogo() {
	logo := `
 _    _    _            _____  _____  _______
| |  / |  | |          / ____||  __ \|__   __|
| | / /| |_| |__   ___| |  __ | |__) |  | |   
| |/ / | __| '_ \ / _ \ | |_ ||  ___/   | |   
|   <  | |_| |_) |  __/ |__| || |       | |   
|_|\_\  \__|_.__/ \___|\_____|_|       |_|   
                                             
`
	color.New(color.FgCyan, color.Bold).Println(logo)
	color.New(color.FgWhite).Println("AI-powered Kubernetes troubleshooting assistant")
	color.New(color.FgWhite).Println("--------------------------------------------")
	fmt.Println()
}