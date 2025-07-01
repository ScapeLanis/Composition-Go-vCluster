package vcluster

import (
	"fmt"

	v1beta1 "github.com/crossplane/crossplane/apis/apiextensions/v1beta1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ScapeLanis/GoVCluster/structs"
)

func CreateRuntimeUsage(Ofresource runtime.Object, Byresource runtime.Object, reason string) (usageobject *v1beta1.Usage, err error) {
	of := Ofresource.DeepCopyObject()
	uof, ok := of.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("failed to cast Ofresource to *unstructured.Unstructured")
	}
	ofname := uof.GetName()
	ofkind := uof.GetKind()
	ofapi := uof.GetAPIVersion()

	by := Byresource.DeepCopyObject()
	uby, ok := by.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("failed to cast By resource to *unstructured.Unstructured")
	}
	bykind := uby.GetKind()
	byapi := uby.GetAPIVersion()
	byname := uby.GetName()
	bylabels := uby.GetLabels()

	usage := &v1beta1.Usage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Usage",
			APIVersion: "apiextensions.crossplane.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ofname + "usage",
		},
		Spec: v1beta1.UsageSpec{
			Of: v1beta1.Resource{
				APIVersion: ofapi,
				Kind:       ofkind,
				ResourceRef: &v1beta1.ResourceRef{
					Name: ofname,
				},
			},
			By: &v1beta1.Resource{
				Kind:       bykind,
				APIVersion: byapi,
				ResourceRef: &v1beta1.ResourceRef{
					Name: byname,
				},
				ResourceSelector: &v1beta1.ResourceSelector{
					MatchLabels:        bylabels,
					MatchControllerRef: structs.BoolPtr(false),
				},
			},
			Reason:         structs.StrPtr(reason),
			ReplayDeletion: structs.BoolPtr(true),
		},
	}

	return usage, nil
}

func CreateObjectUsage(Ofresource *resource.DesiredComposed, Byresource runtime.Object, reason string) (usageDesired *resource.DesiredComposed, err error) {

	ofunstructured := Ofresource.Resource.Unstructured

	ofname := ofunstructured.GetName()
	ofkind := ofunstructured.GetKind()
	ofapi := ofunstructured.GetAPIVersion()

	by := Byresource.DeepCopyObject()
	uby, ok := by.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("failed to cast By resource to *unstructured.Unstructured")
	}
	bykind := uby.GetKind()
	byapi := uby.GetAPIVersion()
	byname := uby.GetName()
	bylabels := uby.GetLabels()

	usage := &v1beta1.Usage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Usage",
			APIVersion: "apiextensions.crossplane.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "usage-" + ofname + "-by-" + byname,
		},
		Spec: v1beta1.UsageSpec{
			Of: v1beta1.Resource{
				APIVersion: ofapi,
				Kind:       ofkind,
				ResourceRef: &v1beta1.ResourceRef{
					Name: ofname,
				},
			},
			By: &v1beta1.Resource{
				Kind:       bykind,
				APIVersion: byapi,
				ResourceRef: &v1beta1.ResourceRef{
					Name: byname,
				},
				ResourceSelector: &v1beta1.ResourceSelector{
					MatchLabels:        bylabels,
					MatchControllerRef: structs.BoolPtr(false),
				},
			},
			Reason:         structs.StrPtr(reason),
			ReplayDeletion: structs.BoolPtr(true),
		},
	}
	o, err := composed.From(usage)
	if err != nil {
		return nil, fmt.Errorf("failed to convert object to composed: %w", err)
	}
	return &resource.DesiredComposed{
		Resource: o,
		Ready:    resource.ReadyTrue,
	}, nil
}

func CreateUsages(desired map[resource.Name]*resource.DesiredComposed, clustername string) error {

	usage_statefulset := &v1beta1.Usage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Usage",
			APIVersion: "apiextensions.crossplane.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "usage-statefulset" + clustername,
		},
		Spec: v1beta1.UsageSpec{
			Of: v1beta1.Resource{
				APIVersion: "kubernetes.crossplane.io/v1alpha2",
				Kind:       "Object",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "statefulset-" + clustername,
				},
			},
			By: &v1beta1.Resource{
				Kind:       "ProviderConfig",
				APIVersion: "kubernetes.crossplane.io/v1alpha1",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "vclusterconfig-" + clustername,
				},
				ResourceSelector: &v1beta1.ResourceSelector{
					MatchLabels: map[string]string{
						"providerforvcluster": "true",
						"app":                 "vcluster-" + clustername,
						"release":             clustername,
					},
					MatchControllerRef: structs.BoolPtr(false),
				},
			},
			Reason:         structs.StrPtr("Ressource im Cluster noch vorhanden"),
			ReplayDeletion: structs.BoolPtr(true),
		},
	}

	a, err := composed.From(usage_statefulset)
	if err != nil {
		return fmt.Errorf("failed to convert object to composed: %w", err)
	}
	desired[resource.Name("usage-statefulset"+clustername)] = &resource.DesiredComposed{
		Resource: a,
		Ready:    resource.ReadyTrue,
	}
	usage_configmap := &v1beta1.Usage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Usage",
			APIVersion: "apiextensions.crossplane.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "usage-configmap-" + clustername,
		},
		Spec: v1beta1.UsageSpec{
			Of: v1beta1.Resource{
				APIVersion: "kubernetes.crossplane.io/v1alpha2",
				Kind:       "Object",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "configmap-" + clustername,
				},
			},
			By: &v1beta1.Resource{
				Kind:       "ProviderConfig",
				APIVersion: "kubernetes.crossplane.io/v1alpha1",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "vclusterconfig-" + clustername,
				},
				ResourceSelector: &v1beta1.ResourceSelector{
					MatchLabels: map[string]string{
						"providerforvcluster": "true",
						"app":                 "vcluster-" + clustername,
						"release":             clustername,
					},
					MatchControllerRef: structs.BoolPtr(false),
				},
			},
			Reason:         structs.StrPtr("Ressource im Clutser noch vorhanden"),
			ReplayDeletion: structs.BoolPtr(true),
		},
	}

	z, err := composed.From(usage_configmap)
	if err != nil {
		return fmt.Errorf("failed to convert object to composed: %w", err)
	}
	desired[resource.Name("usage-configmap-"+clustername)] = &resource.DesiredComposed{
		Resource: z,
		Ready:    resource.ReadyTrue,
	}
	/*
		usage_service := &v1beta1.Usage{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Usage",
				APIVersion: "apiextensions.crossplane.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "usage-service-nodeport-" + clustername,
			},
			Spec: v1beta1.UsageSpec{
				Of: v1beta1.Resource{
					APIVersion: "kubernetes.crossplane.io/v1alpha2",
					Kind:       "Object",
					ResourceRef: &v1beta1.ResourceRef{
						Name: "exposevclusternodeport-" + clustername,
					},
				},
				By: &v1beta1.Resource{
					Kind:       "ProviderConfig",
					APIVersion: "kubernetes.crossplane.io/v1alpha1",
					ResourceRef: &v1beta1.ResourceRef{
						Name: "vclusterconfig-" + clustername,
					},
					ResourceSelector: &v1beta1.ResourceSelector{
						MatchLabels: map[string]string{
							"providerforvcluster": "true",
							"app":                 "vcluster-" + clustername,
							"release":             clustername,
						},
						MatchControllerRef: structs.BoolPtr(false),
					},
				},
				Reason:         structs.StrPtr("Ressource im Clutser noch vorhanden"),
				ReplayDeletion: structs.BoolPtr(true),
			},
		}

		c, err := composed.From(usage_service)
		if err != nil {
			return fmt.Errorf("failed to convert object to composed: %w", err)
		}
		desired[resource.Name("usage-service-"+clustername)] = &resource.DesiredComposed{
			Resource: c,
			Ready:    resource.ReadyTrue,
		}
	*/
	usage_service_vcluster := &v1beta1.Usage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Usage",
			APIVersion: "apiextensions.crossplane.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "usage-service-vcluster-" + clustername,
		},
		Spec: v1beta1.UsageSpec{
			Of: v1beta1.Resource{
				APIVersion: "kubernetes.crossplane.io/v1alpha2",
				Kind:       "Object",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "service-" + clustername,
				},
			},
			By: &v1beta1.Resource{
				Kind:       "ProviderConfig",
				APIVersion: "kubernetes.crossplane.io/v1alpha1",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "vclusterconfig-" + clustername,
				},
				ResourceSelector: &v1beta1.ResourceSelector{
					MatchLabels: map[string]string{
						"providerforvcluster": "true",
						"app":                 "vcluster-" + clustername,
						"release":             clustername,
					},
					MatchControllerRef: structs.BoolPtr(false),
				},
			},
			Reason:         structs.StrPtr("Ressource im Clutser noch vorhanden"),
			ReplayDeletion: structs.BoolPtr(true),
		},
	}

	d, err := composed.From(usage_service_vcluster)
	if err != nil {
		return fmt.Errorf("failed to convert object to composed: %w", err)
	}
	desired[resource.Name("usage-service-vcluster-"+clustername)] = &resource.DesiredComposed{
		Resource: d,
		Ready:    resource.ReadyTrue,
	}
	usage_service_headless := &v1beta1.Usage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Usage",
			APIVersion: "apiextensions.crossplane.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "usage-service-headless-" + clustername,
		},
		Spec: v1beta1.UsageSpec{
			Of: v1beta1.Resource{
				APIVersion: "kubernetes.crossplane.io/v1alpha2",
				Kind:       "Object",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "service-headless-" + clustername,
				},
			},
			By: &v1beta1.Resource{
				Kind:       "ProviderConfig",
				APIVersion: "kubernetes.crossplane.io/v1alpha1",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "vclusterconfig-" + clustername,
				},
				ResourceSelector: &v1beta1.ResourceSelector{
					MatchLabels: map[string]string{
						"providerforvcluster": "true",
						"app":                 "vcluster-" + clustername,
						"release":             clustername,
					},
					MatchControllerRef: structs.BoolPtr(false),
				},
			},
			Reason:         structs.StrPtr("Ressource im Clutser noch vorhanden"),
			ReplayDeletion: structs.BoolPtr(true),
		},
	}

	e, err := composed.From(usage_service_headless)
	if err != nil {
		return fmt.Errorf("failed to convert object to composed: %w", err)
	}
	desired[resource.Name("usage-service-headless-"+clustername)] = &resource.DesiredComposed{
		Resource: e,
		Ready:    resource.ReadyTrue,
	}
	usage_secret := &v1beta1.Usage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Usage",
			APIVersion: "apiextensions.crossplane.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "usage-secret-" + clustername,
		},
		Spec: v1beta1.UsageSpec{
			Of: v1beta1.Resource{
				APIVersion: "kubernetes.crossplane.io/v1alpha2",
				Kind:       "Object",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "secret-" + clustername,
				},
			},
			By: &v1beta1.Resource{
				Kind:       "ProviderConfig",
				APIVersion: "kubernetes.crossplane.io/v1alpha1",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "vclusterconfig-" + clustername,
				},
				ResourceSelector: &v1beta1.ResourceSelector{
					MatchLabels: map[string]string{
						"providerforvcluster": "true",
						"app":                 "vcluster-" + clustername,
						"release":             clustername,
					},
					MatchControllerRef: structs.BoolPtr(false),
				},
			},
			Reason:         structs.StrPtr("Ressource im Clutser noch vorhanden"),
			ReplayDeletion: structs.BoolPtr(true),
		},
	}

	g, err := composed.From(usage_secret)
	if err != nil {
		return fmt.Errorf("failed to convert object to composed: %w", err)
	}
	desired[resource.Name("usage-secret-"+clustername)] = &resource.DesiredComposed{
		Resource: g,
		Ready:    resource.ReadyTrue,
	}

	usage_role := &v1beta1.Usage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Usage",
			APIVersion: "apiextensions.crossplane.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "usage-role-" + clustername,
		},
		Spec: v1beta1.UsageSpec{
			Of: v1beta1.Resource{
				APIVersion: "kubernetes.crossplane.io/v1alpha2",
				Kind:       "Object",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "role-" + clustername,
				},
			},
			By: &v1beta1.Resource{
				Kind:       "ProviderConfig",
				APIVersion: "kubernetes.crossplane.io/v1alpha1",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "vclusterconfig-" + clustername,
				},
				ResourceSelector: &v1beta1.ResourceSelector{
					MatchLabels: map[string]string{
						"providerforvcluster": "true",
						"app":                 "vcluster-" + clustername,
						"release":             clustername,
					},
					MatchControllerRef: structs.BoolPtr(false),
				},
			},
			Reason:         structs.StrPtr("Ressource im Clutser noch vorhanden"),
			ReplayDeletion: structs.BoolPtr(true),
		},
	}

	g, err = composed.From(usage_role)
	if err != nil {
		return fmt.Errorf("failed to convert object to composed: %w", err)
	}
	desired[resource.Name("usage-role-"+clustername)] = &resource.DesiredComposed{
		Resource: g,
		Ready:    resource.ReadyTrue,
	}
	usage_rolebinding := &v1beta1.Usage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Usage",
			APIVersion: "apiextensions.crossplane.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "usage-rolebinding-" + clustername,
		},
		Spec: v1beta1.UsageSpec{
			Of: v1beta1.Resource{
				APIVersion: "kubernetes.crossplane.io/v1alpha2",
				Kind:       "Object",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "rolebinding-" + clustername,
				},
			},
			By: &v1beta1.Resource{
				Kind:       "ProviderConfig",
				APIVersion: "kubernetes.crossplane.io/v1alpha1",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "vclusterconfig-" + clustername,
				},
				ResourceSelector: &v1beta1.ResourceSelector{
					MatchLabels: map[string]string{
						"providerforvcluster": "true",
						"app":                 "vcluster-" + clustername,
						"release":             clustername,
					},
					MatchControllerRef: structs.BoolPtr(false),
				},
			},
			Reason:         structs.StrPtr("Ressource im Clutser noch vorhanden"),
			ReplayDeletion: structs.BoolPtr(true),
		},
	}

	g, err = composed.From(usage_rolebinding)
	if err != nil {
		return fmt.Errorf("failed to convert object to composed: %w", err)
	}
	desired[resource.Name("usage-rolebinding-"+clustername)] = &resource.DesiredComposed{
		Resource: g,
		Ready:    resource.ReadyTrue,
	}

	usage_serviceaccount_workload := &v1beta1.Usage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Usage",
			APIVersion: "apiextensions.crossplane.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "usage-serviceaccount-workload-" + clustername,
		},
		Spec: v1beta1.UsageSpec{
			Of: v1beta1.Resource{
				APIVersion: "kubernetes.crossplane.io/v1alpha2",
				Kind:       "Object",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "serviceaccount-workload-" + clustername,
				},
			},
			By: &v1beta1.Resource{
				Kind:       "ProviderConfig",
				APIVersion: "kubernetes.crossplane.io/v1alpha1",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "vclusterconfig-" + clustername,
				},
				ResourceSelector: &v1beta1.ResourceSelector{
					MatchLabels: map[string]string{
						"providerforvcluster": "true",
						"app":                 "vcluster-" + clustername,
						"release":             clustername,
					},
					MatchControllerRef: structs.BoolPtr(false),
				},
			},
			Reason:         structs.StrPtr("Ressource im Clutser noch vorhanden"),
			ReplayDeletion: structs.BoolPtr(true),
		},
	}

	g, err = composed.From(usage_serviceaccount_workload)
	if err != nil {
		return fmt.Errorf("failed to convert object to composed: %w", err)
	}
	desired[resource.Name("usage-serviceaccount-workload-"+clustername)] = &resource.DesiredComposed{
		Resource: g,
		Ready:    resource.ReadyTrue,
	}
	/*
		usage_serviceaccount := &v1beta1.Usage{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Usage",
				APIVersion: "apiextensions.crossplane.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "usage-serviceaccount-" + clustername,
			},
			Spec: v1beta1.UsageSpec{
				Of: v1beta1.Resource{
					APIVersion: "kubernetes.crossplane.io/v1alpha2",
					Kind:       "Object",
					ResourceRef: &v1beta1.ResourceRef{
						Name: "serviceaccount-vc-" + clustername,
					},
				},
				By: &v1beta1.Resource{
					Kind:       "ProviderConfig",
					APIVersion: "kubernetes.crossplane.io/v1alpha1",
					ResourceRef: &v1beta1.ResourceRef{
						Name: "vclusterconfig-" + clustername,
					},
					ResourceSelector: &v1beta1.ResourceSelector{
						MatchLabels: map[string]string{
							"providerforvcluster": "true",
							"app":                 "vcluster-" + clustername,
							"release":             clustername,
						},
						MatchControllerRef: structs.BoolPtr(false),
					},
				},
				Reason:         structs.StrPtr("Ressource im Clutser noch vorhanden"),
				ReplayDeletion: structs.BoolPtr(true),
			},
		}

		g, err = composed.From(usage_serviceaccount)
		if err != nil {
			return fmt.Errorf("failed to convert object to composed: %w", err)
		}
		desired[resource.Name("usage-serviceaccount-"+clustername)] = &resource.DesiredComposed{
			Resource: g,
			Ready:    resource.ReadyTrue,
		}*/

	//Kubeconfig Secret for vcluster
	usage_secret_kubeconfig := &v1beta1.Usage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Usage",
			APIVersion: "apiextensions.crossplane.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "usage-secret-kubeconfig" + clustername,
		},
		Spec: v1beta1.UsageSpec{
			Of: v1beta1.Resource{
				APIVersion: "kubernetes.crossplane.io/v1alpha2",
				Kind:       "Object",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "vclusterkubeconfig-" + clustername,
				},
			},
			By: &v1beta1.Resource{
				Kind:       "ProviderConfig",
				APIVersion: "kubernetes.crossplane.io/v1alpha1",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "vclusterconfig-" + clustername,
				},
				ResourceSelector: &v1beta1.ResourceSelector{
					MatchLabels: map[string]string{
						"providerforvcluster": "true",
						"app":                 "vcluster-" + clustername,
						"release":             clustername,
					},
					MatchControllerRef: structs.BoolPtr(false),
				},
			},
			Reason:         structs.StrPtr("Ressource im Clutser noch vorhanden"),
			ReplayDeletion: structs.BoolPtr(true),
		},
	}

	g, err = composed.From(usage_secret_kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to convert object to composed: %w", err)
	}
	desired[resource.Name("usage-secret-kubeconfig-"+clustername)] = &resource.DesiredComposed{
		Resource: g,
		Ready:    resource.ReadyTrue,
	}

	usage_statefulset_vclustersecret := &v1beta1.Usage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Usage",
			APIVersion: "apiextensions.crossplane.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "usage-statefulset" + clustername,
		},
		Spec: v1beta1.UsageSpec{
			Of: v1beta1.Resource{
				APIVersion: "kubernetes.crossplane.io/v1alpha2",
				Kind:       "Object",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "statefulset-" + clustername,
				},
			},
			By: &v1beta1.Resource{
				Kind:       "Secret",
				APIVersion: "v1",
				ResourceRef: &v1beta1.ResourceRef{
					Name: "vc-" + clustername,
				},
				/*
					ResourceSelector: &v1beta1.ResourceSelector{
						MatchLabels: map[string]string{
							"app":           "vcluster",
							"vcluster-name": clustername,
						},
						MatchControllerRef: structs.BoolPtr(false),
					},*/
			},
			Reason:         structs.StrPtr("Statefulset" + clustername + "noch vorhanden erst dann Secret löschen"),
			ReplayDeletion: structs.BoolPtr(true),
		},
	}

	a, err = composed.From(usage_statefulset_vclustersecret)
	if err != nil {
		return fmt.Errorf("failed to convert object to composed: %w", err)
	}
	desired[resource.Name("usage-statefulset"+clustername)] = &resource.DesiredComposed{
		Resource: a,
		Ready:    resource.ReadyTrue,
	}

	return nil
}
