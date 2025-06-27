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
	vcluster.ExposeNodePort(req, desired, namespace, clustername, ipadresse)

	resource := req.Observed.Resources["exposevclusternodeport-"+clustername]
	if resource == nil || resource.Resource == nil {
		fmt.Println("resource or resource.Resource is nil")
	} else {
		statusField, ok := resource.Resource.Fields["status"]
		if !ok || statusField == nil {
			fmt.Println("status field not found or nil")
		} else {
			status := statusField.GetStructValue()
			if status == nil {
				fmt.Println("status struct is nil")
			} else {
				atProviderField, ok := status.Fields["atProvider"]
				if !ok || atProviderField == nil {
					fmt.Println("atProvider field not found or nil")
				} else {
					atProvider := atProviderField.GetStructValue()
					if atProvider == nil {
						fmt.Println("atProvider struct is nil")
					} else {
						manifestField, ok := atProvider.Fields["manifest"]
						if !ok || manifestField == nil {
							fmt.Println("manifest field not found or nil")
						} else {
							manifest := manifestField.GetStructValue()
							if manifest == nil {
								fmt.Println("manifest struct is nil")
							} else {
								manifestSpecField, ok := manifest.Fields["spec"]
								if !ok || manifestSpecField == nil {
									fmt.Println("manifest.spec field not found or nil")
								} else {
									manifestSpec := manifestSpecField.GetStructValue()
									if manifestSpec == nil {
										fmt.Println("manifestSpec struct is nil")
									} else {
										portsField, ok := manifestSpec.Fields["ports"]
										if !ok || portsField == nil || portsField.GetListValue() == nil {
											fmt.Println("ports field not found or nil")
										} else {
											ports := portsField.GetListValue().Values
											for _, portVal := range ports {
												if portVal == nil {
													continue
												}
												port := portVal.GetStructValue()
												if nameField, ok := port.Fields["name"]; !ok || nameField.GetStringValue() != "https" {
													continue
												}
												if nodePortField, ok := port.Fields["nodePort"]; ok && nodePortField != nil {
													nodePortStr := fmt.Sprintf("%.0f", nodePortField.GetNumberValue())
													fmt.Println("nodePort:", nodePortStr)

													err := vcluster.Vclustercomponents(req, desired, namespace, clustername, ipadresse, nodePortStr)
													if err != nil {
														response.ConditionFalse(rsp, "FunctionSuccess", "InternalError").
															WithMessage("VClusterComponents failed").
															TargetCompositeAndClaim()
														response.Fatal(rsp, errors.Wrapf(err, "Can not create VClusterComponents"))
													}
													break // Wenn ein gültiger nodePort gefunden wurde, abbrechen
												} else {
													fmt.Println("nodePort not found in port named 'https'")
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
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
