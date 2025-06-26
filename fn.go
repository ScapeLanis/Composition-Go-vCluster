package main

import (
	"context"
	"fmt"

	invcluster "github.com/ScapeLanis/GoVCluster/resourcesinvcluster"
	vcluster "github.com/ScapeLanis/GoVCluster/vCluster"
	objectv1alpha2 "github.com/crossplane-contrib/provider-kubernetes/apis/object/v1alpha2"
	v1beta1 "github.com/crossplane/crossplane/apis/apiextensions/v1beta1"
	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"
	appsv1 "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// Function implements the FunctionRunnerServiceServer.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer
	log logging.Logger
}

func init() {
	//structs.AddToScheme(composed.Scheme)

	composed.Scheme.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.ServiceAccount{})
	composed.Scheme.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.Secret{})
	composed.Scheme.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.ConfigMap{})
	composed.Scheme.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.Service{})
	composed.Scheme.AddKnownTypes(rbacv1.SchemeGroupVersion, &rbacv1.Role{})
	composed.Scheme.AddKnownTypes(rbacv1.SchemeGroupVersion, &rbacv1.RoleBinding{})
	composed.Scheme.AddKnownTypes(appsv1.SchemeGroupVersion, &appsv1.StatefulSet{})
	composed.Scheme.AddKnownTypes(objectv1alpha2.SchemeGroupVersion, &objectv1alpha2.Object{})
	composed.Scheme.AddKnownTypes(v1beta1.SchemeGroupVersion, &v1beta1.Usage{})
	composed.Scheme.AddKnownTypes(batch.SchemeGroupVersion, &batch.Job{})
}

// RunFunction ist der Einstiegspunkt für die Crossplane-Funktion.
func (f *Function) RunFunction(ctx context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)
	//Nodeport and IPAdresse from Minikube VM
	var ipadresse string = "192.168.49.2"
	//var nodeport string = ":30180"
	//Get XR
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
	//Declare Desired
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

	targetName := "vcluster-nodeport-" + clustername
	for _, res := range req.Observed.Resources {
		if res.Resource == nil {
			continue
		}

		kind := res.Resource.Fields["kind"].GetStringValue()
		if kind != "Service" {
			continue
		}

		metadata := res.Resource.Fields["metadata"].GetStructValue()
		name := metadata.Fields["name"].GetStringValue()
		if name != targetName {
			continue
		}

		// Zugriff auf spec.ports[].nodePort aus observed
		spec := res.Resource.Fields["spec"].GetStructValue()
		ports := spec.Fields["ports"].GetListValue().Values

		for _, p := range ports {
			portStruct := p.GetStructValue()
			if npField, ok := portStruct.Fields["nodePort"]; ok {
				nodePort := npField.GetNumberValue()
				fmt.Printf("✅ NodePort of service %q is: %.0f\n", name, nodePort)
			}
		}
	}

	//Create vCluster Components Nodeport entfernt
	err = vcluster.Vclustercomponents(req, desired, namespace, clustername, ipadresse, "tesdfsdfasd")
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "Can not Create VClusterComponents"))
		return rsp, nil
	}
	//Create Usages, Order to Delete in vCluster then vCluster Components
	err = vcluster.CreateUsages(desired, clustername)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "Can not Create Usages"))
		return rsp, nil
	}
	//Create TestPod in vCluster
	/*
		err = invcluster.CreateTestPod(desired, clustername)
		if err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "Can not Create Testpod in vCluster"))
			return rsp, nil
		}*/
	err = invcluster.CreateCertManager(desired, clustername)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "Can not Create Cert-Manager"))
		return rsp, nil
	}

	// Übergib die Desired Ressourcen an die Response
	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot set desired composed resources"))
		return rsp, nil
	}

	return rsp, nil
}
