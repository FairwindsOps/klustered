package cmd

import (
	r "github.com/fairwindsops/klustered/pkg/register"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploys the necessary resources to get the klustered pod running in the cluster.",
	Long:  `Deploys a service account, clusterrolebinding, and a pod for klustered.`,
	Run: func(cmd *cobra.Command, args []string) {
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
