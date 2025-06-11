package main

import (
	"context"
	"fmt"

	v1beta1 "github.com/crossplane/crossplane/apis/apiextensions/v1beta1"

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ScapeLanis/GoVCluster/structs"
	vcluster "github.com/ScapeLanis/GoVCluster/vCluster"
)

// Function implements the FunctionRunnerServiceServer.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer
	log logging.Logger
}

func createObject(obj runtime.Object, name, clustername, providername string) (*resource.DesiredComposed, error) {
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
	composed.Scheme.AddKnownTypes(v1beta1.SchemeGroupVersion, &v1beta1.Usage{})
}

// RunFunction ist der Einstiegspunkt für die Crossplane-Funktion.
func (f *Function) RunFunction(ctx context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	var ipadresse string = "192.168.49.2"
	var nodeport string = ":30180"

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

	rsp, err = vcluster.Vclustercomponents(rsp, req, desired, namespace, clustername, ipadresse, nodeport)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "Can not Create VClusterComponents"))
		return rsp, nil
	}

	err = vcluster.CreateUsages(desired, clustername)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "Can not Create Usages"))
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
			Labels: map[string]string{
				"invcluster": "true",
			},
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
	desired[resource.Name("testpodzwei-"+clustername)], err = createObject(testpodzwei, "testpodzwei-", clustername, "vclusterconfig-"+clustername)
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
