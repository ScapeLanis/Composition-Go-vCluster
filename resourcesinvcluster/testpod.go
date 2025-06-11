package incluster

import (
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createTestPod() {
	testpodzwei := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testpodinvclusterzweitest",
			Namespace: "default",
			Labels: map[string]string{
				"invcluster": "true",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "testcontainer",
					Image: "nginx",
				},
			},
		},
	}
	desired[resource.Name("testpodzwei-"+clustername)], err = createObject(testpodzwei, "testpodzwei-", clustername, "vclusterconfig-"+clustername)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}
}
