package cmd

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"html"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	Short: "Runs the klustered program for breaking the cluster.",
	Long: `Run sets up an http server on 8443 that does a few things:
	
- Responds to a mutating admission webhook on /mutate
- Responds to a validating admissino webhook on /admission
- Responds to health checks on /
- Creates a service, mutatingadmissionconfiguration, and validatingadmissionconfiguration
- Re-creates itself as a new pod if killed gracefully`,
	Run: func(cmd *cobra.Command, args []string) {
		// Setup the auto-registration of webhooks and service
		watcher, err := r.NewWatcher(crt)
		if err != nil {
			klog.Fatal(err)
		}
		watcher.Run()

		// Source: https://rafallorenz.com/go/handle-signals-to-graceful-shutdown-http-server/
		ctx, cancel := context.WithCancel(context.Background())
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
			BaseContext:    func(_ net.Listener) context.Context { return ctx },
		}

		// Run server
		go func() {
			klog.Info("starting listening on :8443")
			if err := s.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
				// it is fine to use Fatal here because it is not main gorutine
				klog.Fatalf("HTTP server ListenAndServe: %v", err)
			}
		}()

		signalChan := make(chan os.Signal, 1)

		signal.Notify(
			signalChan,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGQUIT,
			syscall.SIGTERM,
		)

		<-signalChan
		klog.Info("shutting down...\n")
		watcher.Shutdown()

		go func() {
			<-signalChan
			klog.Fatal("os.Kill - terminating...\n")
		}()

		gracefullCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()

		if err := s.Shutdown(gracefullCtx); err != nil {
			klog.Infof("shutdown error: %v\n", err)
			defer os.Exit(1)
			cancel()
			return
		} else {
			klog.Infof("gracefully stopped\n")
		}

		cancel()

		defer os.Exit(0)
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
