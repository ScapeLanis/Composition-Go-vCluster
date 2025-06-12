package incluster

import (
	vcluster "github.com/ScapeLanis/GoVCluster/vCluster"
	"github.com/crossplane/function-sdk-go/resource"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateTestPod(desired map[resource.Name]*resource.DesiredComposed, clustername string) error {
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

	obj, err := vcluster.CreateObject(testpodzwei, "testpodzwei-", clustername, "vclusterconfig-"+clustername)
	if err != nil {
		return err
	}
	desired[resource.Name("testpod-"+clustername)] = obj
	return err
}
