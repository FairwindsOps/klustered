package cmd

import (
	r "github.com/fairwindsops/klustered/pkg/register"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Setup the auto-registration of webhooks and service
		watcher, err := r.NewWatcher(crt)
		if err != nil {
			klog.Fatal(err)
		}
		watcher.Deploy()
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
