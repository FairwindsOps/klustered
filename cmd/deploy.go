/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
