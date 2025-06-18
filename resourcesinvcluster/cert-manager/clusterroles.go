package certmanager

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func CreateClusterRolesCertManager(clustername, version string) []runtime.Object {

	clusterrole_cert_manager_cainjector := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
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
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"get", "create", "update", "patch"},
			},
			{
				APIGroups: []string{"admissionregistration.k8s.io"},
				Resources: []string{"validatingwebhookconfigurations", "mutatingwebhookconfigurations"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{"apiregistration.k8s.io"},
				Resources: []string{"apiservices"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{"apiextensions.k8s.io"},
				Resources: []string{"customresourcedefinitions"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
		},
	}

	clusterrole_cert_manager_controller_issuers := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
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
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"issuers", "issuers/status"},
				Verbs:     []string{"update", "patch"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"issuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
		},
	}

	clusterrole_cert_manager_controller_clusterissuers := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
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
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"clusterissuers", "clusterissuers/status"},
				Verbs:     []string{"update", "patch"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"clusterissuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
		},
	}

	clusterrole_cert_manager_controller_certificates := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
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
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates", "certificates/status", "certificaterequests", "certificaterequests/status"},
				Verbs:     []string{"update", "patch"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates", "certificaterequests", "clusterissuers", "issuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates/finalizers", "certificaterequests/finalizers"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"order"},
				Verbs:     []string{"create", "delete", "get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
		},
	}

	clusterrole_cert_manager_controller_orders := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
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
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"orders", "orders/status"},
				Verbs:     []string{"update", "patch"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"orders", "challenges"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"clusterissuers", "issuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges"},
				Verbs:     []string{"create", "delete"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"orders/finalizers"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
		},
	}

	clusterrole_cert_manager_http01_controller_challenges := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-http-01-controller-challenges",
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
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges", "challenges/status"},
				Verbs:     []string{"update", "patch"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"clusterissuers", "issuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges/finalizers"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services"},
				Verbs:     []string{"get", "list", "watch", "create", "delete"},
			},
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"ingresses"},
				Verbs:     []string{"get", "list", "watch", "create", "delete", "update"},
			},
			{
				APIGroups: []string{"route.openshift.io"},
				Resources: []string{"routes/custom-host"},
				Verbs:     []string{"create"},
			},
		},
	}

	clusterrole_cert_manager_dns01_controller_challenges := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-dns-01-controller-challenges",
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
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges", "challenges/status"},
				Verbs:     []string{"update", "patch"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"clusterissuers", "issuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges/finalizers"},
				Verbs:     []string{"update"},
			},
		},
	}
	clusterrole_cert_manager_controller_ingress_shim := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
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
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates", "certificaterequests"},
				Verbs:     []string{"create", "update", "delete"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates", "certificaterequests", "issuers", "clusterissuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"ingresses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"ingresses/finalizers"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"gateway.networking.k8s.io"},
				Resources: []string{"gateways", "httproutes"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"gateway.networking.k8s.io"},
				Resources: []string{"gateways/finalizers", "httproutes/finalizers"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
		},
	}
	clusterrole_cert_manager_cluster_view := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-cluster-view",
			Labels: map[string]string{
				"app":                         "cert-manager",
				"app.kubernetes.io/name":      "cert-manager",
				"app.kubernetes.io/instance":  clustername,
				"app.kubernetes.io/component": "controller",
				"app.kubernetes.io/version":   version,
				"rbac.authorization.k8s.io/aggregate-to-cluster-reader": "true",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"clusterissuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
	clusterrole_cert_manager_view := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-cluster-view",
			Labels: map[string]string{
				"app":                                                   "cert-manager",
				"app.kubernetes.io/name":                                "cert-manager",
				"app.kubernetes.io/instance":                            clustername,
				"app.kubernetes.io/component":                           "controller",
				"app.kubernetes.io/version":                             version,
				"rbac.authorization.k8s.io/aggregate-to-view":           "true",
				"rbac.authorization.k8s.io/aggregate-to-edit":           "true",
				"rbac.authorization.k8s.io/aggregate-to-admin":          "true",
				"rbac.authorization.k8s.io/aggregate-to-cluster-reader": "true",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates", "certificaterequests", "issuers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges", "orders"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}
	clusterrole_cert_manager_edit := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-edit",
			Labels: map[string]string{
				"app":                                          "cert-manager",
				"app.kubernetes.io/name":                       "cert-manager",
				"app.kubernetes.io/instance":                   clustername,
				"app.kubernetes.io/component":                  "controller",
				"app.kubernetes.io/version":                    version,
				"rbac.authorization.k8s.io/aggregate-to-edit":  "true",
				"rbac.authorization.k8s.io/aggregate-to-admin": "true",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates", "certificaterequests", "issuers"},
				Verbs:     []string{"create", "delete", "deletecollection", "patch", "update"},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates/status"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges", "orders"},
				Verbs:     []string{"create", "delete", "deletecollection", "patch", "update"},
			},
		},
	}

	clusterrole_cert_manager_controller_approve_cert_manager_io := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-controller-approve:cert-manager-io",
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
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"signers"},
				Verbs:     []string{"approve"},
				ResourceNames: []string{
					"issuers.cert-manager.io/*",
					"clusterissuers.cert-manager.io/*",
				},
			},
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificates/status"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"acme.cert-manager.io"},
				Resources: []string{"challenges", "orders"},
				Verbs:     []string{"create", "delete", "deletecollection", "patch", "update"},
			},
		},
	}

	clusterrole_cert_manager_controller_certificatesigningrequests := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-controller-certificatesigningrequests",
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
				APIGroups: []string{"certificates.k8s.io"},
				Resources: []string{"certificatesigningrequests"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{"certificates.k8s.io"},
				Resources: []string{"certificatesigningrequests/status"},
				Verbs:     []string{"update", "patch"},
			},
			{
				APIGroups: []string{"certificates.k8s.io"},
				Resources: []string{"signers"},
				Verbs:     []string{"sign"},
				ResourceNames: []string{
					"issuers.cert-manager.io/*",
					"clusterissuers.cert-manager.io/*",
				},
			},
			{
				APIGroups: []string{"authorization.k8s.io"},
				Resources: []string{"subjectaccessreviews"},
				Verbs:     []string{"create"},
			},
		},
	}
	clusterrole_cert_manager_webhook_subjectaccessreviews := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
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
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"authorization.k8s.io"},
				Resources: []string{"subjectaccessreviews"},
				Verbs:     []string{"create"},
			},
		},
	}

	return []runtime.Object{
		clusterrole_cert_manager_cainjector,
		clusterrole_cert_manager_controller_issuers,
		clusterrole_cert_manager_controller_clusterissuers,
		clusterrole_cert_manager_controller_certificates,
		clusterrole_cert_manager_controller_orders,
		clusterrole_cert_manager_http01_controller_challenges,
		clusterrole_cert_manager_dns01_controller_challenges,
		clusterrole_cert_manager_controller_ingress_shim,
		clusterrole_cert_manager_cluster_view,
		clusterrole_cert_manager_view,
		clusterrole_cert_manager_edit,
		clusterrole_cert_manager_controller_approve_cert_manager_io,
		clusterrole_cert_manager_controller_certificatesigningrequests,
		clusterrole_cert_manager_webhook_subjectaccessreviews,
	}

}
