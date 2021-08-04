package register

import (
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Watcher struct {
	Client      *kubernetes.Clientset
	Certificate []byte
	Delay       time.Duration
}

func NewWatcher(certificate []byte) (*Watcher, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	watcher := &Watcher{
		Client:      clientset,
		Certificate: certificate,
		Delay:       10 * time.Second,
	}
	return watcher, nil
}

func (w Watcher) Run() {
	ticker := time.NewTicker(w.Delay)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				w.checkMutatingWebhook()
				w.checkValidatingWebhook()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
