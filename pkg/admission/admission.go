package admission

import (
	"encoding/json"
	"fmt"
	"strings"

	v1beta1 "k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func Admit(body []byte, verbose bool) ([]byte, error) {
	klog.V(3).Infof("recv: %s\n", string(body))

	// unmarshal request into AdmissionReview struct
	admReview := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}
	var err error
	var deploy *appsv1.Deployment

	responseBody := []byte{}
	ar := admReview.Request
	resp := v1beta1.AdmissionResponse{}

	if ar != nil {
		if err := json.Unmarshal(ar.Object.Raw, &deploy); err != nil {
			return nil, fmt.Errorf("unable unmarshal deployment json object %v", err)
		}
	}
	// set response options
	resp.Allowed = true
	resp.UID = ar.UID
	resp.Result = &metav1.Status{
		Code: 200,
	}
	for _, container := range deploy.Spec.Template.Spec.Containers {
		if !strings.Contains(container.Image, "klustered:v2") {
			klog.V(6).Infof("skipping image %s", container.Image)
			continue
		}

		klog.V(3).Infof("found container image: %s", container.Image)
		resp.Allowed = false
		resp.Result = &metav1.Status{
			Code:    403,
			Message: "Sorry Carta. <3 Fairwinds",
		}
	}
	admReview.Response = &resp
	// back into JSON so we can return the finished AdmissionReview w/ Response directly
	// w/o needing to convert things in the http handler
	responseBody, err = json.Marshal(admReview)
	if err != nil {
		return nil, err
	}
	klog.V(3).Infof("resp: %s\n", string(responseBody))
	return responseBody, nil
}
