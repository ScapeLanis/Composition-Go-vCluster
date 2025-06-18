package incluster

import (
	certmanager "github.com/ScapeLanis/GoVCluster/resourcesinvcluster/cert-manager"
	vcluster "github.com/ScapeLanis/GoVCluster/vCluster"
	"github.com/crossplane/function-sdk-go/resource"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateCertManager(desired map[resource.Name]*resource.DesiredComposed, clustername string) error {
	namespace := "cert-manager"
	version := "v1.18.0"
	certmanager.CreateClusterRolesCertManager(clustername, version)
	certmanager.CreateServiceAccountsCertManager(namespace, clustername, version)
	namespace_certmanager := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"name":       namespace,
				"invcluster": "true",
			},
		},
	}

	obj, err := vcluster.CreateObject(namespace_certmanager, "namespace-certmanager-", clustername, "vclusterconfig-"+clustername)
	if err != nil {
		return err
	}
	desired[resource.Name("namespace-certmanager-"+clustername)] = obj

	return err
}
