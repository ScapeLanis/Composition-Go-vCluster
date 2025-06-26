package incluster

import (
	"fmt"
	"strings"

	certmanager "github.com/ScapeLanis/GoVCluster/resourcesinvcluster/cert-manager"
	vcluster "github.com/ScapeLanis/GoVCluster/vCluster"
	"github.com/crossplane/function-sdk-go/resource"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func CreateCertManager(desired map[resource.Name]*resource.DesiredComposed, clustername string) error {
	namespace := "cert-manager"

	//Version Images and Labels
	version := "v1.18.0"
	var allResources []runtime.Object
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
	rescrds, err := certmanager.CreateCRDCertManager(namespace, clustername, version)
	if err != nil {
		return err
	}

	resapicheck := certmanager.CreateApiCheckCertManager(namespace, clustername, version)
	resclusterrolebindings := certmanager.CreateClusterRoleBinding(namespace, clustername, version)
	resclusterroles := certmanager.CreateClusterRolesCertManager(clustername, version)
	resdeployments := certmanager.CreateDeploymentsCertManager(namespace, clustername, version)
	resrolebindings := certmanager.CreateRoleBindingsCertManager(namespace, clustername, version)
	resroles := certmanager.CreateRolesCertManager(namespace, clustername, version)
	resserviceaccounts := certmanager.CreateServiceAccountsCertManager(namespace, clustername, version)
	ressvc := certmanager.CreateServicesCertManager(namespace, clustername, version)
	reswebhook := certmanager.CreateWebhooksCertManager(namespace, clustername, version)

	allResources = append(allResources, namespace_certmanager)
	allResources = append(allResources, resapicheck...)
	allResources = append(allResources, rescrds...)
	allResources = append(allResources, resclusterrolebindings...)
	allResources = append(allResources, resclusterroles...)
	allResources = append(allResources, resdeployments...)
	allResources = append(allResources, resrolebindings...)
	allResources = append(allResources, resroles...)
	allResources = append(allResources, resserviceaccounts...)
	allResources = append(allResources, ressvc...)
	allResources = append(allResources, reswebhook...)

	for _, res := range allResources {
		if metaObj, ok := res.(metav1.Object); ok {

			kindlower := strings.ToLower(res.GetObjectKind().GroupVersionKind().Kind)
			// : entfernen oder durch - ersetzen
			cleanName := strings.ReplaceAll(metaObj.GetName(), ":", "-")
			name := cleanName + "-" + kindlower
			obj, err := vcluster.CreateObject(res, name, clustername, "vclusterconfig-"+clustername)
			if err != nil {
				return fmt.Errorf("failed to create %s: %w", name, err)
			}
			desired[resource.Name(name)] = obj
		} else {
			fmt.Println("Das Objekt enthält keine Metadaten.")
		}
	}

	return nil
}
