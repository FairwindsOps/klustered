// source: https://github.com/alex-leonhardt/k8s-mutate-webhook/blob/master/pkg/mutate/mutate.go
package mutate

import (
	"encoding/json"
	"fmt"
	"strings"

	v1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func Mutate(body []byte, verbose bool) ([]byte, error) {

	klog.V(3).Infof("recv: %s\n", string(body))

	// unmarshal request into AdmissionReview struct
	admReview := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}

	var err error
	var pod *corev1.Pod

	responseBody := []byte{}
	ar := admReview.Request
	resp := v1beta1.AdmissionResponse{}

	if ar != nil {
		// get the Pod object and unmarshal it into its struct, if we cannot, we might as well stop here
		if err := json.Unmarshal(ar.Object.Raw, &pod); err != nil {
			return nil, fmt.Errorf("unable unmarshal pod json object %v", err)
		}
		// set response options
		resp.Allowed = true
		resp.UID = ar.UID
		pT := v1beta1.PatchTypeJSONPatch
		resp.PatchType = &pT // it's annoying that this needs to be a pointer as you cannot give a pointer to a constant?

		// add some audit annotations, helpful to know why a object was modified, maybe (?)
		resp.AuditAnnotations = map[string]string{
			"fairwinds.com/hello": "carta",
		}

		// the actual mutation is done by a string in JSONPatch style, i.e. we don't _actually_ modify the object, but
		// tell K8S how it should modifiy it
		p := []map[string]string{}
		for i, container := range pod.Spec.Containers {
			if !strings.Contains(container.Image, "klustered:v2") {
				klog.V(6).Infof("skipping image %s", container.Image)
				continue
			}

			klog.V(3).Infof("found container image: %s", container.Image)
			patch := map[string]string{
				"op":    "replace",
				"path":  fmt.Sprintf("/spec/containers/%d/image", i),
				"value": "ghcr.io/rawkode/klustered:v1",
			}
			p = append(p, patch)
		}
		// parse the []map into JSON
		resp.Patch, err = json.Marshal(p)

		// Success, of course ;)
		resp.Result = &metav1.Status{
			Status: "Success",
		}

		admReview.Response = &resp
		// back into JSON so we can return the finished AdmissionReview w/ Response directly
		// w/o needing to convert things in the http handler
		responseBody, err = json.Marshal(admReview)
		if err != nil {
			return nil, err
		}
	}

	klog.V(3).Infof("resp: %s\n", string(responseBody))

	return responseBody, nil
}
