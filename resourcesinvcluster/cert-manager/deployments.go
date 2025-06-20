package certmanager

import (
	"github.com/ScapeLanis/GoVCluster/structs"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func CreateDeploymentsCertManager(namespace, clustername, version string) []runtime.Object {
	cert_manager_cainjector := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
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
		Spec: appsv1.DeploymentSpec{
			Replicas: structs.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":      "cainjector",
					"app.kubernetes.io/instance":  clustername,
					"app.kubernetes.io/component": "cainjector",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                         "cainjector",
						"app.kubernetes.io/name":      "cainjector",
						"app.kubernetes.io/instance":  clustername,
						"app.kubernetes.io/component": "cainjector",
						"app.kubernetes.io/version":   version,
					},
					Annotations: map[string]string{
						"prometheus.io/path":   "/metrics",
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "9402",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: clustername + "-cert-manager-cainjector",
					EnableServiceLinks: structs.BoolPtr(false),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: structs.BoolPtr(true),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "cert-manager-cainjector",
							Image:           "quay.io/jetstack/cert-manager-cainjector:" + version,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args: []string{
								"--v=2 --leader-election-namespace=kube-system",
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http-metrics",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 9402,
								},
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
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: structs.BoolPtr(false),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{
										"ALL",
									},
								},
								ReadOnlyRootFilesystem: structs.BoolPtr(true),
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

	cert_manager := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
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
		Spec: appsv1.DeploymentSpec{
			Replicas: structs.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":      "cert-manager",
					"app.kubernetes.io/instance":  clustername,
					"app.kubernetes.io/component": "controller",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                         "cert-manager",
						"app.kubernetes.io/name":      "cert-manager",
						"app.kubernetes.io/instance":  clustername,
						"app.kubernetes.io/component": "controller",
						"app.kubernetes.io/version":   version,
					},
					Annotations: map[string]string{
						"prometheus.io/path":   "/metrics",
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "9402",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: clustername + "-cert-manager",
					EnableServiceLinks: structs.BoolPtr(false),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: structs.BoolPtr(true),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "cert-manager-controller",
							Image:           "quay.io/jetstack/cert-manager-controller:" + version,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args: []string{
								"--v=2 --cluster-resource-namespace=$(POD_NAMESPACE) --leader-election-namespace=kube-system --acme-http01-solver-image=quay.io/jetstack/cert-manager-acmesolver:" + version + " --max-concurrent-challenges=60",
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http-metrics",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 9402,
								},
								{
									Name:          "http-healthz",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 9403,
								},
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
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: structs.BoolPtr(false),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{
										"ALL",
									},
								},
								ReadOnlyRootFilesystem: structs.BoolPtr(true),
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/livez",
										Port:   intstr.FromString("http-healthz"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       10,
								TimeoutSeconds:      15,
								SuccessThreshold:    1,
								FailureThreshold:    8,
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

	cert_manager_webhook := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
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
		Spec: appsv1.DeploymentSpec{
			Replicas: structs.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":      "webhook",
					"app.kubernetes.io/instance":  clustername,
					"app.kubernetes.io/component": "webhook",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                         "webhook",
						"app.kubernetes.io/name":      "webhook",
						"app.kubernetes.io/instance":  clustername,
						"app.kubernetes.io/component": "webhook",
						"app.kubernetes.io/version":   version,
					},
					Annotations: map[string]string{
						"prometheus.io/path":   "/metrics",
						"prometheus.io/scrape": "true",
						"prometheus.io/port":   "9402",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: clustername + "-cert-manager-webhook",
					EnableServiceLinks: structs.BoolPtr(false),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: structs.BoolPtr(true),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "cert-manager-webhook",
							Image:           "quay.io/jetstack/cert-manager-webhook:" + version,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args: []string{
								"--v=2 --secure-port=10250 --dynamic-serving-ca-secret-namespace=$(POD_NAMESPACE) --dynamic-serving-ca-secret-name=" + clustername + "-cert-manager-webhook --dynamic-serving-dns-names=" + clustername + "-cert-manager-webhook.$(POD_NAMESPACE) --dynamic-serving-dns-names=" + clustername + "-cert-manager-webhook.$(POD_NAMESPACE).svc",
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "https",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 10250,
								},
								{
									Name:          "healthcheck",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 6080,
								},
								{
									Name:          "http-metrics",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 9402,
								},
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
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: structs.BoolPtr(false),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{
										"ALL",
									},
								},
								ReadOnlyRootFilesystem: structs.BoolPtr(true),
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/livez",
										Port:   intstr.FromString("healthcheck"),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 60,
								PeriodSeconds:       10,
								TimeoutSeconds:      1,
								SuccessThreshold:    1,
								FailureThreshold:    3,
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
		cert_manager_cainjector,
		cert_manager,
		cert_manager_webhook,
	}
}
