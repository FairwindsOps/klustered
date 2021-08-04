package register

import (
	"context"

	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (w Watcher) checkService() error {
	hooks, err := w.Client.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, svc := range hooks.Items {
		if svc.Name == "api-server" {
			klog.V(3).Infof("found service : %s", svc.Name)
			return err
		}
	}
	klog.V(3).Infof("no service found")
	err = w.createService()
	if err != nil {
		klog.Error(err)
	}
	return nil
}

func (w Watcher) createService() error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-server",
			Namespace: "kube-system",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "ssl",
					Port:       8443,
					TargetPort: intstr.FromInt(8443),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"component": "cilium-api",
			},
		},
	}

	_, err := w.Client.CoreV1().Services("kube-system").Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
