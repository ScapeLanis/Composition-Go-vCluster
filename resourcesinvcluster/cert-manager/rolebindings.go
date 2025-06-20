package certmanager

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func CreateRoleBindingsCertManager(namespace, clustername, version string) []runtime.Object {
	cert_manager_cainjector_leaderelection := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
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
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     clustername + "-cert-manager-cainjector:leaderelection",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      clustername + "-cert-manager-cainjector",
				Namespace: namespace,
				Kind:      "ServiceAccount",
			},
		},
	}

	cert_manager_leaderelection := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
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
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     clustername + "-cert-manager:leaderelection",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      clustername + "-cert-manager",
				Namespace: namespace,
				Kind:      "ServiceAccount",
			},
		},
	}

	cert_manager_tokenrequest := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
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
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     clustername + "-cert-manager-tokenrequest",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      clustername + "-cert-manager",
				Namespace: namespace,
				Kind:      "ServiceAccount",
			},
		},
	}

	cert_manager_webhook_dynamic_serving := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
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
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     clustername + "-cert-manager-webhook:dynamic-serving",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      clustername + "-cert-manager-webhook",
				Namespace: namespace,
				Kind:      "ServiceAccount",
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
