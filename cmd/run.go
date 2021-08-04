/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"crypto/tls"
	_ "embed"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"time"

	a "github.com/fairwindsops/klustered/pkg/admission"
	m "github.com/fairwindsops/klustered/pkg/mutate"
	r "github.com/fairwindsops/klustered/pkg/register"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

//go:embed ssl/api-server.pem
var crt []byte

//go:embed ssl/api-server.key
var key []byte

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		watcher, err := r.NewWatcher(crt)
		if err != nil {
			klog.Fatal(err)
		}
		watcher.Run()

		mux := http.NewServeMux()

		mux.HandleFunc("/", handleRoot)
		mux.HandleFunc("/mutate", handleMutate)
		mux.HandleFunc("/admission", handleAdmission)

		// Generate a key pair from your pem-encoded cert and key ([]byte).
		cert, err := tls.X509KeyPair(crt, key)
		if err != nil {
			klog.Fatal(err)
		}

		// Construct a tls.config
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		s := &http.Server{
			Addr:           ":8443",
			Handler:        mux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20, // 1048576,
			TLSConfig:      tlsConfig,
		}

		klog.Info("starting listening on :8443")
		klog.Fatal(s.ListenAndServeTLS("", ""))
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello %q", html.EscapeString(r.URL.Path))
}

func handleAdmission(w http.ResponseWriter, r *http.Request) {
	klog.V(3).Info("handling /admission request")
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		klog.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
	}
	admission, err := a.Admit(body, true)
	if err != nil {
		klog.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
	}

	// and write it back
	w.WriteHeader(http.StatusOK)
	w.Write(admission)
}

func handleMutate(w http.ResponseWriter, r *http.Request) {
	klog.V(3).Info("handling /mutate request")
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		klog.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
	}

	// mutate the request
	mutated, err := m.Mutate(body, true)
	if err != nil {
		klog.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
	}

	// and write it back
	w.WriteHeader(http.StatusOK)
	w.Write(mutated)
}
