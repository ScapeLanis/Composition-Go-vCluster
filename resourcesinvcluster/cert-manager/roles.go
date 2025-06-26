package certmanager

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func CreateRolesCertManager(namespace, clustername, version string) []runtime.Object {

	cert_manager_cainjector_leaderelection := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername + "-cert-manager-cainjector:leaderelection",
			Namespace: "kube-system",

			Labels: map[string]string{
				"app":                         "cainjector",
				"app.kubernetes.io/name":      "cainjector",
				"app.kubernetes.io/instance":  clustername,
				"app.kubernetes.io/component": "cainjector",
				"app.kubernetes.io/version":   version,
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"coordination.k8s.io"},
				Resources:     []string{"leases"},
				ResourceNames: []string{"cert-manager-cainjector-leader-election", "cert-manager-cainjector-leader-election-core"},
				Verbs:         []string{"get", "update", "patch"},
			},
			{
				APIGroups: []string{"coordination.k8s.io"},
				Resources: []string{"leases"},
				Verbs:     []string{"create"},
			},
		},
	}

	cert_manager_leaderelection := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername + "-cert-manager:leaderelection",
			Namespace: "kube-system",

			Labels: map[string]string{
				"app":                         "cert-manager",
				"app.kubernetes.io/name":      "cert-manager",
				"app.kubernetes.io/instance":  clustername,
				"app.kubernetes.io/component": "controller",
				"app.kubernetes.io/version":   version,
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"coordination.k8s.io"},
				Resources:     []string{"leases"},
				ResourceNames: []string{"cert-manager-controller"},
				Verbs:         []string{"get", "update", "patch"},
			},
			{
				APIGroups: []string{"coordination.k8s.io"},
				Resources: []string{"leases"},
				Verbs:     []string{"create"},
			},
		},
	}

	cert_manager_tokenrequest := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername + "-cert-manager-tokenrequest",
			Namespace: namespace,

			Labels: map[string]string{
				"app":                         "cert-manager",
				"app.kubernetes.io/name":      "cert-manager",
				"app.kubernetes.io/instance":  clustername,
				"app.kubernetes.io/component": "controller",
				"app.kubernetes.io/version":   version,
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{""},
				Resources:     []string{"leaseserviceaccounts/tokens"},
				ResourceNames: []string{clustername + "-cert-manager"},
				Verbs:         []string{"create"},
			},
		},
	}

	cert_manager_webhook_dynamic_serving := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername + "-cert-manager-webhook:dynamic-serving",
			Namespace: namespace,

			Labels: map[string]string{
				"app":                         "webhook",
				"app.kubernetes.io/name":      "webhook",
				"app.kubernetes.io/instance":  clustername,
				"app.kubernetes.io/component": "webhook",
				"app.kubernetes.io/version":   version,
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{""},
				Resources:     []string{"secrets"},
				ResourceNames: []string{clustername + "-cert-manager-webhook-ca"},
				Verbs:         []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"create"},
			},
		},
	}
	return []runtime.Object{
		cert_manager_cainjector_leaderelection,
		cert_manager_leaderelection,
		cert_manager_tokenrequest,
		cert_manager_webhook_dynamic_serving,
	}
}
