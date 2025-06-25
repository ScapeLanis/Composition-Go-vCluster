package certmanager

import (
	"github.com/ScapeLanis/GoVCluster/structs"
	batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func CreateApiCheckCertManager(namespace, clustername, version string) []runtime.Object {
	serviceaccount_startupapicheck := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername + "-cert-manager-startupapicheck",
			Namespace: namespace,
			Labels: map[string]string{
				"app":                         "startupapicheck",
				"app.kubernetes.io/instance":  "clustername",
				"app.kubernetes.io/component": "startupapicheck",
				"app.kubernetes.io/version":   version,
			},
		},
	}
	role_startupapicheck := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername + "-startupapicheck:create-cert",
			Namespace: namespace,
			Labels: map[string]string{
				"app":                         "startupapicheck",
				"app.kubernetes.io/instance":  "clustername",
				"app.kubernetes.io/component": "startupapicheck",
				"app.kubernetes.io/version":   version,
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"cert-manager.io"},
				Resources: []string{"certificaterequests"},
				Verbs:     []string{"create"},
			},
		},
	}

	rolebinding_startupapicheck := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername + "-cert-manager-startupapicheck:create-cert",
			Namespace: namespace,
			Labels: map[string]string{
				"app":                         "startupapicheck",
				"app.kubernetes.io/name":      "startupapicheck",
				"app.kubernetes.io/instance":  clustername,
				"app.kubernetes.io/component": "startupapicheck",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     role_startupapicheck.Kind,
			Name:     role_startupapicheck.Name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      serviceaccount_startupapicheck.Kind,
				Name:      serviceaccount_startupapicheck.Name,
				Namespace: namespace,
			},
		},
	}

	job_apicheck := &batch.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername + "-cert-manager-startupapicheck",
			Namespace: namespace,
			Labels: map[string]string{
				"app":                         "startupapicheck",
				"app.kubernetes.io/name":      "startupapicheck",
				"app.kubernetes.io/instance":  clustername,
				"app.kubernetes.io/component": "startupapicheck",
			},
		},
		Spec: batch.JobSpec{
			BackoffLimit: structs.Int32Ptr(4),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                         "startupapicheck",
						"app.kubernetes.io/name":      "startupapicheck",
						"app.kubernetes.io/instance":  clustername,
						"app.kubernetes.io/component": "startupapicheck",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					ServiceAccountName: serviceaccount_startupapicheck.Name,
					EnableServiceLinks: structs.BoolPtr(false),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: structs.BoolPtr(true),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "cert-manager-startupapicheck",
							Image:           "quay.io/jetstack/cert-manager-startupapicheck:" + version,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args:            []string{"check", "api", "--wait=1m", "-v"},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: structs.BoolPtr(false),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{
										"All",
									},
								},
								ReadOnlyRootFilesystem: structs.BoolPtr(true),
							},
							Env: []corev1.EnvVar{
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					NodeSelector: map[string]string{
						"kubernetes.io/os": "linux",
					},
				},
			},
		},
	}

	return []runtime.Object{

		role_startupapicheck,
		rolebinding_startupapicheck,
		job_apicheck,
	}
}
