package certmanager

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func CreateClusterRoleBinding(namespace, clustername, version string) []runtime.Object {

	cert_manager_cainjector := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-cainjector",
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
			Kind:     "ClusterRole",
			Name:     clustername + "-cert-manager-cainjector",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      clustername + "-cert-manager-cainjector",
				Namespace: namespace,
				Kind:      "ServiceAccount",
			},
		},
	}
	cert_manager_controller_issuers := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-controller-issuers",
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
			Kind:     "ClusterRole",
			Name:     clustername + "-cert-manager-controller-issuers",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      clustername + "-cert-manager",
				Namespace: namespace,
				Kind:      "ServiceAccount",
			},
		},
	}

	cert_manager_controller_clusterissuers := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-controller-clusterissuers",
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
			Kind:     "ClusterRole",
			Name:     clustername + "-cert-manager-controller-clusterissuers",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      clustername + "-cert-manager",
				Namespace: namespace,
				Kind:      "ServiceAccount",
			},
		},
	}

	cert_manager_controller_certificates := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-controller-certificates",
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
			Kind:     "ClusterRole",
			Name:     clustername + "-cert-manager-controller-certificates",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      clustername + "-cert-manager",
				Namespace: namespace,
				Kind:      "ServiceAccount",
			},
		},
	}

	cert_manager_controller_orders := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-controller-orders",
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
			Kind:     "ClusterRole",
			Name:     clustername + "-cert-manager-controller-orders",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      clustername + "-cert-manager",
				Namespace: namespace,
				Kind:      "ServiceAccount",
			},
		},
	}

	cert_manager_http01_controller_challenges := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-http01-controller-challenges",
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
			Kind:     "ClusterRole",
			Name:     clustername + "-cert-manager-http01-controller-challenges",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      clustername + "-cert-manager",
				Namespace: namespace,
				Kind:      "ServiceAccount",
			},
		},
	}
	cert_manager_dns01_controller_challenges := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-controller-issuers",
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
			Kind:     "ClusterRole",
			Name:     clustername + "-cert-manager-dns01-controller-challenges",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      clustername + "-cert-manager",
				Namespace: namespace,
				Kind:      "ServiceAccount",
			},
		},
	}
	cert_manager_controller_ingress_shim := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-controller-ingress-shim",
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
			Kind:     "ClusterRole",
			Name:     clustername + "-cert-manager-controller-ingress-shim",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      clustername + "-cert-manager",
				Namespace: namespace,
				Kind:      "ServiceAccount",
			},
		},
	}
	cert_manager_controller_approve_cert_manager_io := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-controller-approve:cert-manager-io",
			Labels: map[string]string{
				"app":                         "cert-manager",
				"app.kubernetes.io/name":      "cert-manager",
				"app.kubernetes.io/instance":  clustername,
				"app.kubernetes.io/component": "cert-manager",
				"app.kubernetes.io/version":   version,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clustername + "-cert-manager-controller-approve:cert-manager-io",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      clustername + "-cert-manager",
				Namespace: namespace,
				Kind:      "ServiceAccount",
			},
		},
	}

	cert_manager_certificatesigningrequests := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-controller-certificatesigningrequests",
			Labels: map[string]string{
				"app":                         "cert-manager",
				"app.kubernetes.io/name":      "cert-manager",
				"app.kubernetes.io/instance":  clustername,
				"app.kubernetes.io/component": "cert-manager",
				"app.kubernetes.io/version":   version,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clustername + "-cert-manager-controller-certificatesigningrequests",
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      clustername + "-cert-manager",
				Namespace: namespace,
				Kind:      "ServiceAccount",
			},
		},
	}
	cert_manager_webhook_subjectaccessreviews := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-webhook:subjectaccessreviews",
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
			Kind:     "ClusterRole",
			Name:     clustername + "-cert-manager-webhook:subjectaccessreviews",
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
		cert_manager_cainjector,
		cert_manager_controller_issuers,
		cert_manager_controller_clusterissuers,
		cert_manager_controller_certificates,
		cert_manager_controller_orders,
		cert_manager_http01_controller_challenges,
		cert_manager_dns01_controller_challenges,
		cert_manager_controller_ingress_shim,
		cert_manager_controller_approve_cert_manager_io,
		cert_manager_certificatesigningrequests,
		cert_manager_webhook_subjectaccessreviews,
	}
}
