package vcluster

import (
	"fmt"
	"strconv"

	objectv1alpha2 "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha2"
	providerv1alpha1 "github.com/crossplane-contrib/provider-kubernetes/apis/v1alpha1"
	kconfig "github.com/crossplane-contrib/provider-kubernetes/pkg/kube/config"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"

	"strings"

	"github.com/ScapeLanis/GoVCluster/structs"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {

	composed.Scheme.AddKnownTypes(providerv1alpha1.SchemeGroupVersion, &providerv1alpha1.ProviderConfig{})

}
func CreateObject(obj runtime.Object, name, clustername, providername string) (*resource.DesiredComposed, error) {
	if obj == nil {
		return nil, fmt.Errorf("runtime.Object input is nil")
	}

	objekt := &objectv1alpha2.Object{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Object",
			APIVersion: "kubernetes.crossplane.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name + clustername,
			Labels: map[string]string{
				"app":     "vcluster-" + clustername,
				"release": clustername,
			},
		},
		Spec: objectv1alpha2.ObjectSpec{
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{
					Name: providername,
				},
			},
			ForProvider: objectv1alpha2.ObjectParameters{
				Manifest: runtime.RawExtension{
					Object: obj,
				},
			},
		},
	}

	o, err := composed.From(objekt)
	if err != nil {
		return nil, fmt.Errorf("failed to convert object to composed: %w", err)
	}
	return &resource.DesiredComposed{
		Resource: o,
		Ready:    resource.ReadyTrue,
	}, nil

}

func Vclustercomponents(req *fnv1.RunFunctionRequest, desired map[resource.Name]*resource.DesiredComposed, namespace, clustername, ipadresse, nodeport string) error {
	cleannodeport := strings.TrimPrefix(nodeport, ":")
	portInt64, err := strconv.ParseInt(cleannodeport, 10, 32)
	if err != nil {
		fmt.Println("Fehler beim Parsen:", err)
		return err
	}

	nodeportint32 := int32(portInt64)

	serviceaccount_vc := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vc-" + clustername,
			Namespace: namespace,
			Labels: map[string]string{
				"app":     "vcluster",
				"release": clustername,
			},
		},
	}
	obj, err := CreateObject(serviceaccount_vc, "serviceaccount-vc-", clustername, "kubernetes-provider")
	if err != nil {
		return err
	}
	desired[resource.Name("serviceaccount_vc-"+clustername)] = obj

	serviceaccount_workload := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vc-workload-" + clustername,
			Namespace: namespace,
			Labels: map[string]string{
				"app":     "vcluster",
				"release": clustername,
			},
		},
	}
	obj, err = CreateObject(serviceaccount_workload, "serviceaccount-workload-", clustername, "kubernetes-provider")
	if err != nil {
		return err
	}
	desired[resource.Name("serviceaccount-workload-"+clustername)] = obj

	vclusterconfig, err := structs.NewDefaultConfig()
	if err != nil {
		return err
	}

	vclusterconfig.ControlPlane.Proxy.ExtraSANs = append(
		vclusterconfig.ControlPlane.Proxy.ExtraSANs,
		ipadresse,
	)

	yamlBytes, err := yaml.Marshal(vclusterconfig)
	if err != nil {
		return err
	}

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vc-config-" + clustername,
			Namespace: namespace,
			Labels: map[string]string{
				"app":     "vcluster",
				"release": clustername,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"config.yaml": yamlBytes,
		},
	}

	obj, err = CreateObject(secret, "secret-", clustername, "kubernetes-provider")
	if err != nil {
		return err
	}
	desired[resource.Name("secret-"+clustername)] = obj

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
	var result string
	dns_objects := []runtime.Object{coredns_clusterrole, coredns_clusterrolebinding, coredns_configmap, coredns_deployment, coredns_serviceaccount, coredns_service}
	for _, dnsobj := range dns_objects {
		data, err := yaml.Marshal(dnsobj)
		if err != nil {
			return err
		}
		result += string(data) + ("\n---\n")
	}

	configmap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vc-coredns-" + clustername,
			Namespace: namespace,
		},
		Data: map[string]string{
			"coredns.yaml": result,
		},
	}
	obj, err = CreateObject(configmap, "configmap-", clustername, "kubernetes-provider")
	if err != nil {
		return err
	}
	desired[resource.Name("configmap-"+clustername)] = obj

	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vc-" + clustername,
			Namespace: namespace,
			Labels: map[string]string{
				"app":     "vcluster",
				"release": clustername,
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps", "secrets", "services", "pods", "pods/attach", "pods/portforward", "pods/exec", "persistentvolumeclaims"},
				Verbs:     []string{"create", "delete", "patch", "update", "get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods/status", "pods/ephemeralcontainers"},
				Verbs:     []string{"patch", "update"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"statefulsets", "replicasets", "deployments"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"endpoints", "events", "pods/log"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"endpoints"},
				Verbs:     []string{"create", "delete", "patch", "update"},
			},
		},
	}
	obj, err = CreateObject(role, "role-", clustername, "kubernetes-provider")
	if err != nil {
		return err
	}
	desired[resource.Name("role-"+clustername)] = obj

	rolebinding := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vc-" + clustername,
			Namespace: namespace,
			Labels: map[string]string{
				"app":     "vcluster",
				"release": clustername,
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "vc-" + clustername,
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "vc-" + clustername,
		},
	}
	obj, err = CreateObject(rolebinding, "rolebinding-", clustername, "kubernetes-provider")
	if err != nil {
		return err
	}
	desired[resource.Name("rolebinding-"+clustername)] = obj

	service_headless := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername + "-headless",
			Namespace: namespace,
			Labels: map[string]string{
				"app":     "vcluster",
				"release": clustername,
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Port:       443,
					Name:       "https",
					TargetPort: intstr.FromInt(8443),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			PublishNotReadyAddresses: true,
			ClusterIP:                "None",
			Selector: map[string]string{
				"app":     "vcluster",
				"release": clustername,
			},
		},
	}
	obj, err = CreateObject(service_headless, "service-headless-", clustername, "kubernetes-provider")
	if err != nil {
		return err
	}
	desired[resource.Name("service-headless-"+clustername)] = obj

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername,
			Namespace: namespace,
			Labels: map[string]string{
				"app":                      "vcluster",
				"release":                  clustername,
				"vcluster.loft.sh/service": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Port:       443,
					Name:       "https",
					TargetPort: intstr.FromInt(8443),
					NodePort:   0,
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "kubelet",
					Port:       10250,
					TargetPort: intstr.FromInt(8443),
					NodePort:   0,
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app":     "vcluster",
				"release": clustername,
			},
		},
	}
	obj, err = CreateObject(service, "service-", clustername, "kubernetes-provider")
	if err != nil {
		return err
	}
	desired[resource.Name("service-"+clustername)] = obj

	statefulset := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername,
			Namespace: namespace,
			Labels: map[string]string{
				"app":     "vcluster",
				"release": clustername,
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: structs.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     "vcluster",
					"release": clustername,
				},
			},
			PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: appsv1.RetainPersistentVolumeClaimRetentionPolicyType,
			},
			ServiceName:         clustername + "-headless",
			PodManagementPolicy: appsv1.ParallelPodManagement,
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: k8sresource.MustParse("5Gi"),
							},
						},
					},
				},
			},

			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"vClusterConfigHash": "b1768483e2256a4f33a31821c0a9122b283e532dd7decbd7c361caf4540066ec",
					},
					Labels: map[string]string{
						"app":     "vcluster",
						"release": clustername,
					},
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: structs.Int64Ptr(10),
					ServiceAccountName:            "vc-" + clustername,
					Volumes: []corev1.Volume{
						{
							Name: "helm-cache",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "binaries",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "tmp",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "certs",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "vcluster-config",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "vc-config-" + clustername,
								},
							},
						},
						{
							Name: "coredns",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "vc-coredns-" + clustername,
									},
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:  "vcluster-copy",
							Image: "ghcr.io/loft-sh/vcluster-pro:0.24.1",
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/binaries",
									Name:      "binaries",
								},
							},
							Command:         []string{"/bin/sh"},
							Args:            []string{"-c", "cp /vcluster /binaries/vcluster"},
							SecurityContext: &corev1.SecurityContext{},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    k8sresource.MustParse("100m"),
									corev1.ResourceMemory: k8sresource.MustParse("256Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    k8sresource.MustParse("40m"),
									corev1.ResourceMemory: k8sresource.MustParse("64Mi"),
								},
							},
						},
						{
							Name:  "kube-controller-manager",
							Image: "registry.k8s.io/kube-controller-manager:v1.32.0",
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/binaries",
									Name:      "binaries",
								},
							},
							Command:         []string{"/binaries/vcluster"},
							Args:            []string{"cp", "/usr/local/bin/kube-controller-manager", "/binaries/kube-controller-manager"},
							SecurityContext: &corev1.SecurityContext{},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    k8sresource.MustParse("100m"),
									corev1.ResourceMemory: k8sresource.MustParse("256Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    k8sresource.MustParse("40m"),
									corev1.ResourceMemory: k8sresource.MustParse("64Mi"),
								},
							},
						},
						{
							Name:  "kube-apiserver",
							Image: "registry.k8s.io/kube-apiserver:v1.32.0",
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/binaries",
									Name:      "binaries",
								},
							},
							Command:         []string{"/binaries/vcluster"},
							Args:            []string{"cp", "/usr/local/bin/kube-apiserver", "/binaries/kube-apiserver"},
							SecurityContext: &corev1.SecurityContext{},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    k8sresource.MustParse("100m"),
									corev1.ResourceMemory: k8sresource.MustParse("256Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    k8sresource.MustParse("40m"),
									corev1.ResourceMemory: k8sresource.MustParse("64Mi"),
								},
							},
						},
					},
					EnableServiceLinks: structs.BoolPtr(true),
					Containers: []corev1.Container{
						{
							Name:  "syncer",
							Image: "ghcr.io/loft-sh/vcluster-pro:0.24.1",
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/healthz",
										Port:   intstr.FromInt(8443),
										Scheme: corev1.URISchemeHTTPS,
									},
								},
								FailureThreshold:    60,
								InitialDelaySeconds: 60,
								TimeoutSeconds:      3,
								PeriodSeconds:       2,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/readyz",
										Port:   intstr.FromInt(8443),
										Scheme: corev1.URISchemeHTTPS,
									},
								},
								FailureThreshold: 60,
								TimeoutSeconds:   3,
								PeriodSeconds:    2,
							},
							StartupProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/readyz",
										Port:   intstr.FromInt(8443),
										Scheme: corev1.URISchemeHTTPS,
									},
								},
								FailureThreshold: 300,
								TimeoutSeconds:   3,
								PeriodSeconds:    6,
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: structs.BoolPtr(false),
								RunAsGroup:               structs.Int64Ptr(0),
								RunAsUser:                structs.Int64Ptr(0),
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceEphemeralStorage: k8sresource.MustParse("8Gi"),
									corev1.ResourceMemory:           k8sresource.MustParse("2Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:              k8sresource.MustParse("200m"),
									corev1.ResourceEphemeralStorage: k8sresource.MustParse("400Mi"),
									corev1.ResourceMemory:           k8sresource.MustParse("256Mi"),
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "VCLUSTER_NAME",
									Value: clustername,
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: "POD_IP",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.podIP",
										},
									},
								},
								{
									Name: "NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/data",
								},
								{
									Name:      "binaries",
									MountPath: "/binaries",
								},
								{
									Name:      "certs",
									MountPath: "/pki",
								},
								{
									Name:      "helm-cache",
									MountPath: "/.cache/helm",
								},
								{
									Name:      "vcluster-config",
									MountPath: "/var/vcluster",
								},
								{
									Name:      "tmp",
									MountPath: "/tmp",
								},
								{
									Name:      "coredns",
									MountPath: "/manifests/coredns",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
		},
	}
	statefulsetobj := &objectv1alpha2.Object{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Object",
			APIVersion: "kubernetes.crossplane.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "statefulset-" + clustername,
			Labels: map[string]string{
				"app":     "vcluster-" + clustername,
				"release": clustername,
			},
		},
		Spec: objectv1alpha2.ObjectSpec{
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{
					Name: "kubernetes-provider",
				},
				WriteConnectionSecretToReference: &xpv1.SecretReference{
					Name:      "vc-kubeconfig-provider-" + clustername,
					Namespace: namespace,
				},
			},
			ForProvider: objectv1alpha2.ObjectParameters{
				Manifest: runtime.RawExtension{
					Object: statefulset,
				},
			},
			ConnectionDetails: []objectv1alpha2.ConnectionDetail{
				{
					ObjectReference: corev1.ObjectReference{
						APIVersion: "v1",
						Kind:       "Secret",
						Namespace:  namespace,
						Name:       "vc-" + clustername,
						FieldPath:  "data.config",
					},
					ToConnectionSecretKey: "config",
				},
			},
		},
	}
	so, err := composed.From(statefulsetobj)
	if err != nil {
		return err
	}
	desired[resource.Name("statefulset-"+clustername)] = &resource.DesiredComposed{
		Resource: so,
		Ready:    resource.ReadyTrue,
	}

	//ProviderConfig for vCluster
	providerconfig := &providerv1alpha1.ProviderConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ProviderConfig",
			APIVersion: "kubernetes.crossplane.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "vclusterconfig-" + clustername,
			Labels: map[string]string{
				"providerforvcluster": "true",
				"app":                 "vcluster-" + clustername,
				"release":             clustername,
			},
		},
		Spec: kconfig.ProviderConfigSpec{
			Credentials: kconfig.ProviderCredentials{
				Source: "Secret",
				CommonCredentialSelectors: xpv1.CommonCredentialSelectors{
					SecretRef: &xpv1.SecretKeySelector{

						Key: "config.yaml",
						SecretReference: xpv1.SecretReference{
							Namespace: namespace,
							Name:      "vc-kubeconfig-" + clustername,
						},
					},
				},
			},
		},
	}

	u, err := composed.From(providerconfig)
	if err != nil {
		return fmt.Errorf("failed to convert object to composed: %w", err)
	}
	desired[resource.Name("providerconfig-"+clustername)] = &resource.DesiredComposed{
		Resource: u,
		Ready:    resource.ReadyTrue,
	}

	exposevclusternodeport := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vcluster-nodeport-" + clustername,
			Namespace: namespace,
			Labels: map[string]string{
				"app":     "vcluster",
				"release": clustername,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":     "vcluster",
				"release": clustername,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Port:       443,
					TargetPort: intstr.FromInt(8443),
					Protocol:   corev1.ProtocolTCP,
					NodePort:   nodeportint32,
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}
	obj, err = CreateObject(exposevclusternodeport, "exposevclusternodeport-", clustername, "kubernetes-provider")
	if err != nil {
		return err
	}
	desired[resource.Name("exposevclusternodeport-"+clustername)] = obj

	_, err = CreateUsage(exposevclusternodeport, providerconfig, "weil halt")

	resourcestateful := req.Observed.Resources["statefulset-"+clustername]
	if resourcestateful == nil {
		//err = fmt.Errorf("observed resource statefulset-%s is nil", clustername)
		return nil
	} else {
		connection_details := resourcestateful.GetConnectionDetails()
		if connection_details != nil {
			kubeconfigBytes, ok := connection_details["config"]
			if !ok {
				return err
			}
			kubeconfig, err := clientcmd.Load(kubeconfigBytes)
			if err != nil {
				return err
			}
			cluster, exists := kubeconfig.Clusters["kubernetes"]
			if !exists || cluster == nil {
				return err
			}
			kubeconfig.Clusters["kubernetes"].Server = ipadresse + nodeport
			configbytechanged, err := clientcmd.Write(*kubeconfig)
			if err != nil {
				return err
			}

			secretvcluster_kubeconfig := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vc-kubeconfig-" + clustername,
					Namespace: namespace,
					Labels: map[string]string{
						"app":     "vcluster",
						"release": clustername,
					},
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					"config.yaml": configbytechanged,
				},
			}
			obj, err = CreateObject(secretvcluster_kubeconfig, "vclusterkubeconfig-", clustername, "kubernetes-provider")
			if err != nil {
				return err
			}
			desired[resource.Name("secretvcluster-kubeconfig-"+clustername)] = obj

		}

	}
	return nil
}
