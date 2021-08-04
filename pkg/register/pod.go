package register

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
)

func (w Watcher) createPod() error {
	klog.V(3).Infof("re-creating self")
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "cilium-",
			Namespace:    "kube-system",
			Labels: map[string]string{
				"component": "cilium-api",
				"tier":      "control-plane",
			},
		},
		Spec: corev1.PodSpec{
			PriorityClassName:  "system-node-critical",
			ServiceAccountName: "api-server",
			Containers: []corev1.Container{
				{
					Name:            "cilium-c4r7a",
					Image:           "sudermanjr/klustered:dev",
					ImagePullPolicy: corev1.PullAlways,
					Command:         []string{"/klustered", "run", "-v10"},
					LivenessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   "/",
								Port:   intstr.FromInt(8443),
								Scheme: corev1.URISchemeHTTPS,
							},
						},
					},
					ReadinessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path:   "/",
								Port:   intstr.FromInt(8443),
								Scheme: corev1.URISchemeHTTPS,
							},
						},
					},
				},
			},
		},
	}

	_, err := w.Client.CoreV1().Pods("kube-system").Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
