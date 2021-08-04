package register

import (
	"context"

	"k8s.io/klog/v2"

	AdmissionregistrationV1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (w Watcher) checkValidatingWebhook() error {
	hooks, err := w.Client.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, hook := range hooks.Items {
		if hook.Name == "default" {
			klog.V(3).Infof("found admission webhook: %s", hook.Name)
			return err
		}
	}
	klog.V(3).Infof("no admission webhook found")
	err = w.createValidatingWebhook()
	if err != nil {
		klog.Error(err)
	}
	return nil
}

func (w Watcher) createValidatingWebhook() error {
	port := int32(8443)
	path := "/admission"
	sideEffects := AdmissionregistrationV1.SideEffectClassNone
	var timeout int32 = 5
	failurePolicy := AdmissionregistrationV1.FailurePolicyType("Fail")
	validatingHook := &AdmissionregistrationV1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "kube-system",
		},
		Webhooks: []AdmissionregistrationV1.ValidatingWebhook{
			{
				Name:                    "api-server.kube-system.svc.cluster.local",
				AdmissionReviewVersions: []string{"v1beta1", "v1"},
				ClientConfig: AdmissionregistrationV1.WebhookClientConfig{
					Service: &AdmissionregistrationV1.ServiceReference{
						Namespace: "kube-system",
						Name:      "api-server",
						Path:      &path,
						Port:      &port,
					},
					CABundle: w.Certificate,
				},
				Rules: []AdmissionregistrationV1.RuleWithOperations{
					{
						Operations: []AdmissionregistrationV1.OperationType{"CREATE", "UPDATE"},
						Rule: AdmissionregistrationV1.Rule{
							APIGroups:   []string{"apps"},
							APIVersions: []string{"v1"},
							Resources:   []string{"deployments"},
						},
					},
				},
				SideEffects:    &sideEffects,
				TimeoutSeconds: &timeout,
				FailurePolicy:  &failurePolicy,
			},
		},
	}
	_, err := w.Client.AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(context.TODO(), validatingHook, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
