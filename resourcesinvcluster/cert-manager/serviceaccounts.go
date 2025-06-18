package certmanager

import (
	"github.com/ScapeLanis/GoVCluster/structs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func CreateServiceAccountsCertManager(namespace, clustername, version string) []runtime.Object {

	serviceaccount_cainjector := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername + "-cert-manager-cainjector",
			Namespace: namespace,
			Labels: map[string]string{
				"app":                         "cainjector",
				"app.kubernetes.io/name":      "cainjector",
				"app.kubernetes.io/instance":  clustername,
				"app.kubernetes.io/component": "cainjector",
				"app.kubernetes.io/version":   version,
			},
		},
		AutomountServiceAccountToken: structs.BoolPtr(true),
	}

	serviceaccount_cert_manager := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername + "-cert-manager",
			Namespace: namespace,
			Labels: map[string]string{
				"app":                         "cert-manager",
				"app.kubernetes.io/name":      "cert-manager",
				"app.kubernetes.io/instance":  clustername,
				"app.kubernetes.io/component": "controller",
				"app.kubernetes.io/version":   version,
			},
		},
		AutomountServiceAccountToken: structs.BoolPtr(true),
	}

	serviceaccount_webhook := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername + "-cert-manager-webhook",
			Namespace: namespace,
			Labels: map[string]string{
				"app":                         "webhook",
				"app.kubernetes.io/name":      "webhook",
				"app.kubernetes.io/instance":  clustername,
				"app.kubernetes.io/component": "webhook",
				"app.kubernetes.io/version":   version,
			},
		},
		AutomountServiceAccountToken: structs.BoolPtr(true),
	}
	return []runtime.Object{
		serviceaccount_cainjector,
		serviceaccount_cert_manager,
		serviceaccount_webhook,
	}
}
