package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	sigsyaml "sigs.k8s.io/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"
)

// Function implements the FunctionRunnerServiceServer.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer
	log logging.Logger
}



// RunFunction ist der Einstiegspunkt für die Crossplane-Funktion.
func (f *Function) RunFunction(_ context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
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
	
	// Lade Ressourcen aus externer YAML-Datei
	resources, err := loadConfig("config.yaml")
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot load config.yaml"))
		return rsp, nil
	}

	// Iteriere über alle Ressourcen aus YAML
	for i, u := range resources {


		kind, found, err := unstructured.NestedString(u.Object, "kind")
		if err != nil {
			response.Fatal(rsp, errors.Wrap(err, "cannot read metadata.namespace"))
			return rsp, nil
		}
		if found && kind == "StatefulSet" {
			if err := setConnectionDetails(u, clustername, namespace); err != nil{
				response.Fatal(rsp, errors.Wrap(err, "cannoct set connection details"))
				return rsp, nil
			}
		}

		//Namespace auslesen
		ns, found, err := unstructured.NestedString(u.Object, "metadata", "namespace")
		if err != nil {
			response.Fatal(rsp, errors.Wrap(err, "cannot read metadata.namespace"))
			return rsp, nil
		}
		//Namespace setzen wenn vorhanden und default ist
		if found && ns == "default" {
			if err := unstructured.SetNestedField(u.Object, namespace, "metadata", "namespace"); err != nil {
				response.Fatal(rsp, errors.Wrap(err, "cannot set new namespace"))
				return rsp, nil
			}
		}
		// Name setzen
		if err := unstructured.SetNestedField(u.Object, clustername, "metadata", "name"); err != nil {
			response.Fatal(rsp, errors.Wrap(err, "cannot set clustername label"))
			return rsp, nil
		}

		// ProviderConfig setzen
		if err := setProviderConfig(u, "kubernetes-provider"); err != nil {
			response.Fatal(rsp, errors.Wrap(err, "cannot set providerConfigRef"))
			return rsp, nil
		}

		

		// Füge zur DesiredMap hinzu
		desired[resource.Name(fmt.Sprintf("example-resource-%d", i))] = &resource.DesiredComposed{
			Resource: u,
			Ready:    resource.ReadyTrue,
		}
	}

	// Übergib die Desired Ressourcen an die Response
	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot set desired composed resources"))
		return rsp, nil
	}

	return rsp, nil
}

// setProviderConfig fügt ein ProviderConfigRef-Feld hinzu
func setProviderConfig(u *composed.Unstructured, providerName string) error {
	return unstructured.SetNestedField(u.Object, providerName, "spec", "providerConfigRef", "name")
}

// loadConfig lädt eine YAML-Datei mit mehreren Ressourcen
func loadConfig(path string) ([]*composed.Unstructured, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML: %w", err)
	}

	docs := strings.Split(string(data), "---")
	var result []*composed.Unstructured

	for _, doc := range docs {
		if strings.TrimSpace(doc) == "" {
			continue
		}

		var obj map[string]interface{}
		if err := sigsyaml.Unmarshal([]byte(doc), &obj); err != nil {
			return nil, fmt.Errorf("error unmarshalling YAML: %w", err)
		}

		c := composed.New()
		c.SetUnstructuredContent(obj)

		result = append(result, c)
	}

	return result, nil
}
//gibt eine methode in composed mit setConnectionDetails
func setConnectionDetails(u *composed.Unstructured, namespace string, name string) error {
	// connectionDetails ist ein Slice deswegen []mit interface vor der map
	secret := []interface{}{
		map[string]interface{}{
			"apiVersion":            "v1",
			"kind":                  "Secret",
			"name":                  "vc-" + name,
			"namespace":             namespace,
			"fieldPath":             "data.config",
			"toConnectionSecretKey": "kubeconfig",
		},
	}

	if err := unstructured.SetNestedField(u.Object, secret, "spec", "connectionDetails"); err != nil {
		return err
	}

	// writeConnectionSecretToRef ist ein einzelnes Map-Objekt
	newSecret := map[string]interface{}{
		"name":      "kubeconfig-provider-" + name,
		"namespace": namespace,
	}

	if err := unstructured.SetNestedField(u.Object, newSecret, "spec", "writeConnectionSecretToRef"); err != nil {
		return err
	}

	return nil
}

//Kubeconfig vom Secret verändern
//ProviderConfig mit dem ConnectionSecret vom Statefulset
//connectionDetails weg lassen secret mit gleichen namen erstellen auf managementpolicy observed??


/*
// SetWriteConnectionSecretToReference of this Composed resource.
func (cd *Unstructured) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	_ = fieldpath.Pave(cd.Object).SetValue("spec.writeConnectionSecretToRef", r)
}

// GetPublishConnectionDetailsTo of this Composed resource.
func (cd *Unstructured) GetPublishConnectionDetailsTo() *xpv1.PublishConnectionDetailsTo {
	out := &xpv1.PublishConnectionDetailsTo{}
	if err := fieldpath.Pave(cd.Object).GetValueInto("spec.publishConnectionDetailsTo", out); err != nil {
		return nil
	}
	return out
}

// SetPublishConnectionDetailsTo of this Composed resource.
func (cd *Unstructured) SetPublishConnectionDetailsTo(ref *xpv1.PublishConnectionDetailsTo) {
	_ = fieldpath.Pave(cd.Object).SetValue("spec.publishConnectionDetailsTo", ref)
}

// GetValue of the supplied field path.
func (cd *Unstructured) GetValue(path string) (any, error) {
	return fieldpath.Pave(cd.Object).GetValue(path)
}*/