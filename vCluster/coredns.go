package vcluster

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"

	"github.com/ScapeLanis/GoVCluster/structs"
)

func Createcorednscomponents() (result string, err error) {

	//CoreDNS Deployment mit Objekten erstellen und in die ConfigMap einfügen
	coredns_serviceaccount := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "coredns",
			Namespace: "kube-system",
		},
	}

	coredns_clusterrole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:coredns",
			Labels: map[string]string{
				"kubernetes.io/bootstrapping": "rbac-defaults",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"endpoints", "services", "pods", "namespaces"},
				Verbs:     []string{"list", "watch"},
			},
			{
				APIGroups: []string{"discovery.k8s.io"},
				Resources: []string{"endpointslices"},
				Verbs:     []string{"list", "watch"},
			},
		},
	}
	coredns_clusterrolebinding := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "system:coredns",
			Annotations: map[string]string{
				"rbac.authorization.kubernetes.io/autoupdate": "true",
			},
			Labels: map[string]string{
				"kubernetes.io/bootstrapping": "rbac-defaults",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "system:coredns",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "coredns",
				Namespace: "kube-system",
			},
		},
	}
	coredns_configmap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "coredns",
			Namespace: "kube-system",
		},
		Data: map[string]string{
			"Corefile": `
    .:1053 {
        errors
        health
        ready
        rewrite name regex .*\.nodes\.vcluster\.com kubernetes.default.svc.cluster.local
        kubernetes cluster.local in-addr.arpa ip6.arpa {
            pods insecure
            fallthrough in-addr.arpa ip6.arpa
        }
        hosts /etc/NodeHosts {
            ttl 60
            reload 15s
            fallthrough
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        loop
        loadbalance
    }
  
    import /etc/coredns/custom/*.server`,
			"NodeHosts": "",
		},
	}
	coredns_deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "coredns",
			Namespace: "kube-system",
			Labels: map[string]string{
				"k8s-app":            "vcluster-kube-dns",
				"kubernetes.io/name": "CoreDNS",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: structs.Int32Ptr(1),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: structs.IntOrStrPtr(1),
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"k8s-app": "vcluster-kube-dns",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"k8s-app": "vcluster-kube-dns",
					},
				},
				Spec: corev1.PodSpec{
					PriorityClassName:  "",
					ServiceAccountName: "coredns",
					NodeSelector: map[string]string{
						"kubernetes.io/os": "linux",
					},
					TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
						{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"k8s-app": "vcluster-kube-dns",
								},
							},
							MaxSkew:           1,
							TopologyKey:       "kubernetes.io/hostname",
							WhenUnsatisfiable: corev1.DoNotSchedule,
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "coredns",
							Image:           "registry.k8s.io/coredns/coredns:v1.12.0",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    k8sresource.MustParse("200m"),
									corev1.ResourceMemory: k8sresource.MustParse("170Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    k8sresource.MustParse("20m"),
									corev1.ResourceMemory: k8sresource.MustParse("64Mi"),
								},
							},
							Args: []string{"-conf", "/etc/coredns/Corefile"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config-volume",
									MountPath: "/etc/coredns",
									ReadOnly:  true,
								},
								{
									Name:      "custom-config-volume",
									MountPath: "/etc/coredns/custom",
									ReadOnly:  true,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								RunAsUser:                structs.Int64Ptr(100),
								RunAsGroup:               structs.Int64Ptr(100),
								AllowPrivilegeEscalation: structs.BoolPtr(false),
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{
										"NET_BIND_SERVICE",
									},
									Drop: []corev1.Capability{
										"ALL",
									},
								},
								ReadOnlyRootFilesystem: structs.BoolPtr(true),
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/health",
										Port:   intstr.FromInt32(8000),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 60,
								PeriodSeconds:       10,
								TimeoutSeconds:      1,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/ready",
										Port:   intstr.FromInt32(8181),
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 0,
								PeriodSeconds:       2,
								TimeoutSeconds:      1,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
						},
					},
					DNSPolicy: corev1.DNSDefault,
					Volumes: []corev1.Volume{
						{
							Name: "config-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "coredns",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "Corefile",
											Path: "Corefile",
										},
										{
											Key:  "NodeHosts",
											Path: "NodeHosts",
										},
									},
								},
							},
						},
						{
							Name: "custom-config-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "coredns-custom",
									},
									Optional: structs.BoolPtr(true),
								},
							},
						},
					},
				},
			},
		},
	}

	coredns_service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-dns",
			Namespace: "kube-system",
			Annotations: map[string]string{
				"prometheus.io/port":   "9153",
				"prometheus.io/scrape": "true",
			},
			Labels: map[string]string{
				"k8s-app":                       "vcluster-kube-dns",
				"kubernetes.io/cluster-service": "true",
				"kubernetes.io/name":            "CoreDNS",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Port:       53,
					Name:       "dns",
					TargetPort: intstr.FromInt(1053),
					Protocol:   corev1.ProtocolUDP,
				},
				{
					Name:       "dns-tcp",
					Port:       53,
					TargetPort: intstr.FromInt(1053),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:     "metrics",
					Port:     9153,
					Protocol: corev1.ProtocolSCTP,
				},
			},
			Selector: map[string]string{
				"k8s-app": "vcluster-kube-dns",
			},
		},
	}
	dns_objects := []runtime.Object{coredns_clusterrole, coredns_clusterrolebinding, coredns_configmap, coredns_deployment, coredns_serviceaccount, coredns_service}
	for _, dnsobj := range dns_objects {
		data, err := yaml.Marshal(dnsobj)
		if err != nil {
			return "", err
		}
		result += string(data) + ("\n---\n")
	}
	return result, nil
}
