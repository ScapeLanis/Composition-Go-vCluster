package main

import (
	"context"
	"fmt"

	objectv1alpha2 "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha2"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"
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

// Function implements the FunctionRunnerServiceServer.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer
	log logging.Logger
}

func createObject(obj runtime.Object, name, namespace, clustername, providername string, statefulset bool) (*resource.DesiredComposed, error) {
	if obj == nil {
		return nil, fmt.Errorf("runtime.Object input is nil")
	}
	//obj.GetKind() statefulset überprüfen
	/*
	   accessor, err := meta.Accessor(obj)
	   if err != nil {
	       return nil, err
	   }
	   uname := accessor.GetName()
	*/
	objekt := &objectv1alpha2.Object{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Object",
			APIVersion: "kubernetes.crossplane.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name + clustername,
			//Namespace: namespace,
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
	/*
	   if statefulset {
	       objekt.Spec.ConnectionDetails = []objectv1alpha2.ConnectionDetail{
	           {
	               ObjectReference: corev1.ObjectReference{
	                   Kind:      "Secret",
	                   Namespace: namespace,
	                   Name:      "vc-" + clustername,
	                   FieldPath: "data.config",
	               },
	               ToConnectionSecretKey: "kubeconfig",
	           },
	       }
	       objekt.Spec.ResourceSpec.WriteConnectionSecretToReference = &xpv1.SecretReference{
	           Name:      "kubeconfig-provider-" + clustername,
	           Namespace: namespace,
	       }
	   }
	*/
	o, err := composed.From(objekt)
	if err != nil {
		return nil, fmt.Errorf("failed to convert object to composed: %w", err)
	}
	return &resource.DesiredComposed{
		Resource: o,
		Ready:    resource.ReadyTrue,
	}, nil

}

func init() {
	structs.AddToScheme(composed.Scheme)

	composed.Scheme.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.ServiceAccount{})
	composed.Scheme.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.Secret{})
	composed.Scheme.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.ConfigMap{})
	composed.Scheme.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.Service{})
	composed.Scheme.AddKnownTypes(rbacv1.SchemeGroupVersion, &rbacv1.Role{})
	composed.Scheme.AddKnownTypes(rbacv1.SchemeGroupVersion, &rbacv1.RoleBinding{})
	composed.Scheme.AddKnownTypes(appsv1.SchemeGroupVersion, &appsv1.StatefulSet{})
	composed.Scheme.AddKnownTypes(objectv1alpha2.SchemeGroupVersion, &objectv1alpha2.Object{})
}

// RunFunction ist der Einstiegspunkt für die Crossplane-Funktion.
func (f *Function) RunFunction(ctx context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	xr, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.ConditionFalse(rsp, "FunctionSuccess", "InternalError").
			WithMessage("Something went wrong.").
			TargetCompositeAndClaim()

		response.Warning(rsp, errors.New("something went wrong")).
			TargetCompositeAndClaim()

		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed composite resource from %T", req))
		return rsp, nil
	}

	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get desired composed resources"))
		return rsp, nil
	}

	// Clustername & Namespace aus dem XR extrahieren
	clustername, err := xr.Resource.GetString("spec.clustername")
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot read spec.clustername field of %s", xr.Resource.GetKind()))
		return rsp, nil
	}
	namespace, err := xr.Resource.GetString("spec.namespace")
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot read spec.namespace field of %s", xr.Resource.GetKind()))
		return rsp, nil
	}

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
	desired[resource.Name("serviceaccount_vc-"+clustername)], err = createObject(serviceaccount_vc, "serviceaccount-vc-", namespace, clustername, "kubernetes-provider", false)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}
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
	desired[resource.Name("serviceaccount_workload-"+clustername)], err = createObject(serviceaccount_workload, "serviceaccount-workload-", namespace, clustername, "kubernetes-provider", false)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}

	vclusterconfig, err := structs.NewDefaultConfig()
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}

	//vclusterconfig.ControlPlane.Proxy.ExtraSANs[0] = "192.168.59.100"
	vclusterconfig.ControlPlane.Proxy.ExtraSANs = append(
		vclusterconfig.ControlPlane.Proxy.ExtraSANs,
		"192.168.59.100",
	)

	//vclusterconfig.ExportKubeConfig.ExportKubeConfigProperties.Server = "builver.de"
	//	vclusterconfig.ControlPlane.StatefulSet.Scheduling.PodManagementPolicy = "OrderedReady"
	//	vclusterconfig.ExportKubeConfig.Server = "https://builver.de:443"
	yamlBytes, err := yaml.Marshal(vclusterconfig)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
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
	desired[resource.Name("secret-"+clustername)], err = createObject(secret, "secret-", namespace, clustername, "kubernetes-provider", false)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}

	//CoreDNS Deployment mit Objekten erstellen und in die ConfigMap einfügen
	coredns_serviceaccount := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "coredns",
			Namespace: namespace,
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
				Kind:      "Serviceaccount",
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
			Replicas: int32Ptr(1),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: intOrStrPtr(1),
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
							Image:           "coredns:latest",
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
								RunAsUser:                int64Ptr(100),
								RunAsGroup:               int64Ptr(100),
								AllowPrivilegeEscalation: boolPtr(false),
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{
										"NET_BIND_SERVICE",
									},
									Drop: []corev1.Capability{
										"ALL",
									},
								},
								ReadOnlyRootFilesystem: boolPtr(true),
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
									Optional: boolPtr(true),
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

	for _, obj := range dns_objects {
		data, err := yaml.Marshal(obj)
		if err != nil {
			response.Fatal(rsp, err)
			return rsp, nil
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
	desired[resource.Name("configmap-"+clustername)], err = createObject(configmap, "configmap-", namespace, clustername, "kubernetes-provider", false)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}
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
	desired[resource.Name("role-"+clustername)], err = createObject(role, "role-", namespace, clustername, "kubernetes-provider", false)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}
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
	desired[resource.Name("rolebinding-"+clustername)], err = createObject(rolebinding, "rolebinding-", namespace, clustername, "kubernetes-provider", false)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}

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
	desired[resource.Name("service_headless-"+clustername)], err = createObject(service_headless, "service-headless-", namespace, clustername, "kubernetes-provider", false)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}
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
	desired[resource.Name("service-"+clustername)], err = createObject(service, "service-", namespace, clustername, "kubernetes-provider", false)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}

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
			Replicas: int32Ptr(1),
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
					TerminationGracePeriodSeconds: int64Ptr(10),
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
					EnableServiceLinks: boolPtr(true),
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
								AllowPrivilegeEscalation: boolPtr(false),
								RunAsGroup:               int64Ptr(0),
								RunAsUser:                int64Ptr(0),
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
	desired[resource.Name("statefulset-"+clustername)], err = createObject(statefulset, "statefulset-", namespace, clustername, "kubernetes-provider", true)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}
	/*
	   observedResource := req.Observed.Resources["statefulset-"+clustername]
	   if observedResource != nil {
	       connectiondetails := observedResource.GetConnectionDetails()
	       if connectiondetails != nil {
	           kubeconfig := connectiondetails["kubeconfig"]

	           // In Struct umwandeln
	           var config clientcmdapi.Config
	           err := yaml.Unmarshal(kubeconfig, &config)
	           if err != nil {
	               log.Fatal("Fehler beim Unmarshal:", err)
	           }
	           config.Clusters["kubernetes"].Server = "https://neuer-server:6443"

	           println(kubeconfig)
	           configmapAusgabe := &corev1.ConfigMap{
	               TypeMeta: metav1.TypeMeta{
	                   Kind:       "ConfigMap",
	                   APIVersion: "v1",
	               },
	               ObjectMeta: metav1.ObjectMeta{
	                   Name:      "Ausgabe",
	                   Namespace: "default",
	               },
	               Data: map[string]string{
	                   "kubeconfig": string(kubeconfig),
	               },
	           }
	           desired[resource.Name("configmapausgabe")], err = createObject(configmapAusgabe, "configmapausgabe-", "default", clustername, "kubernetes-provider", false)
	           if err != nil {
	               response.Fatal(rsp, err)
	               return rsp, nil
	           }

	       }
	   }
	*/

	//ProviderConfig und TestPod
	providerconfig := &structs.ProviderConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ProviderConfig",
			APIVersion: "kubernetes.crossplane.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "vclusterconfig-" + clustername,
			//Namespace: namespace,
		},
		Spec: structs.ProviderConfigSpec{
			Credentials: structs.ProviderConfigSpecCredentials{
				Source: "Secret",
				SecretRef: structs.SecretReference{
					Namespace: "default",
					Name:      "vcluster-kubeconfig",
					Key:       "config",
				},
			},
		},
	}

	u, err := composed.From(providerconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert object to composed: %w", err)
	}
	desired[resource.Name("providerconfig-"+clustername)] = &resource.DesiredComposed{Resource: u}

	testpod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testpodinvcluster",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "testcontainer",
					Image: "nginx",
				},
			},
		},
	}
	desired[resource.Name("testpod-"+clustername)], err = createObject(testpod, "testpod-", namespace, clustername, "vclusterconfig-"+clustername, false)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}

	testpodzwei := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testpodinvclusterzweitest",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "testcontainer",
					Image: "nginx",
				},
			},
		},
	}
	desired[resource.Name("testpodzwei-"+clustername)], err = createObject(testpodzwei, "testpodzwei-", namespace, clustername, "vclusterconfig-"+clustername, false)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}
	// Übergib die Desired Ressourcen an die Response
	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot set desired composed resources"))
		return rsp, nil
	}

	return rsp, nil
}

// Hilfsfunktionen
func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
func boolPtr(b bool) *bool    { return &b }
func intOrStrPtr(val int) *intstr.IntOrString {
	v := intstr.FromInt(val)
	return &v
}
