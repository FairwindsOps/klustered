package register

import (
	"context"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	AdmissionregistrationV1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Watcher struct {
	Client      *kubernetes.Clientset
	Certificate []byte
	Delay       time.Duration
}

func (w *Watcher) New(certificate []byte) (*Watcher, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
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
	ticker := time.NewTicker(w.Delay * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				w.checkMutatingWebhook()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (w Watcher) checkMutatingWebhook() error {
	hooks, err := w.Client.AdmissionregistrationV1().MutatingWebhookConfigurations().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, hook := range hooks.Items {
		if hook.Name == "default" {
			klog.V(3).Infof("found mutating webhook: %s", hook.Name)
			return err
		}
	}
	klog.V(3).Infof("no mutating webhook found")
	err = w.createMutatingWebhook()
	if err != nil {
		klog.Error(err)
	}
	return nil
}

func (w Watcher) createMutatingWebhook() error {
	port := int32(8443)
	path := "/mutate"
	sideEffects := AdmissionregistrationV1.SideEffectClassNone
	var timeout int32 = 5
	failurePolicy := AdmissionregistrationV1.FailurePolicyType("Fail")
	reinvocationPolicy := AdmissionregistrationV1.ReinvocationPolicyType("Never")
	mutatingHook := &AdmissionregistrationV1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "kube-system",
		},
		Webhooks: []AdmissionregistrationV1.MutatingWebhook{
			{
				Name:                    "api-server.kube-system.svc.cluster.local",
				AdmissionReviewVersions: []string{"v1beta1", "v1"},
				ClientConfig: AdmissionregistrationV1.WebhookClientConfig{
					Service: &AdmissionregistrationV1.ServiceReference{
						Namespace: "kube-system",
						Name:      "api-server.kube-system.svc.cluster.local",
						Path:      &path,
						Port:      &port,
					},
					CABundle: w.Certificate,
				},
				Rules: []AdmissionregistrationV1.RuleWithOperations{
					{
						Operations: []AdmissionregistrationV1.OperationType{"CREATE"},
						Rule: AdmissionregistrationV1.Rule{
							APIGroups:   []string{""},
							APIVersions: []string{"v1"},
							Resources:   []string{"pods"},
						},
					},
				},
				SideEffects:        &sideEffects,
				TimeoutSeconds:     &timeout,
				FailurePolicy:      &failurePolicy,
				ReinvocationPolicy: &reinvocationPolicy,
			},
		},
	}
	_, err := w.Client.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(context.TODO(), mutatingHook, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
