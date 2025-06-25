package certmanager

/*
import (
	"fmt"
	"io"
	"os"

	"github.com/crossplane/function-sdk-go/resource/composed"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
)

func loadConfig(path string) ([]*composed.Unstructured, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open YAML file: %w", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	var result []*composed.Unstructured
	var docIndex int

	for {
		var obj map[string]interface{}
		err := decoder.Decode(&obj)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error decoding YAML document #%d: %w", docIndex+1, err)
		}

		docIndex++

		// Skip empty Documents
		if len(obj) == 0 {
			continue
		}

		if obj["apiVersion"] == nil || obj["kind"] == nil {
			// Skip no valid kubernetes objects
			continue
		}

		c := composed.New()
		c.SetUnstructuredContent(obj)

		result = append(result, c)
	}

	return result, nil
}
func CreateCRDCertManager(namespace, clustername, version string) ([]runtime.Object, error) {
	// Lade Ressourcen aus externer YAML-Datei
	resources, err := loadConfig("cert-manager-crds.yaml")
	if err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, r := range resources {
		result = append(result, r) // castet implizit to runtime.Object
	}
	return result, nil
}
*/

import (
	"bytes"
	"fmt"
	"io"

	"github.com/crossplane/function-sdk-go/resource/composed"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"

	_ "embed"
)

//go:embed cert-manager-crds.yaml
var crdsYaml []byte

func loadConfigFromBytes(data []byte) ([]*composed.Unstructured, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var result []*composed.Unstructured
	var docIndex int

	for {
		var obj map[string]interface{}
		err := decoder.Decode(&obj)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error decoding YAML document #%d: %w", docIndex+1, err)
		}

		docIndex++

		// Skip empty Documents
		if len(obj) == 0 {
			continue
		}

		if obj["apiVersion"] == nil || obj["kind"] == nil {
			// Skip invalid kubernetes objects
			continue
		}

		c := composed.New()
		c.SetUnstructuredContent(obj)

		result = append(result, c)
	}

	return result, nil
}

func CreateCRDCertManager(namespace, clustername, version string) ([]runtime.Object, error) {
	// Lade Ressourcen aus eingebetteter YAML
	resources, err := loadConfigFromBytes(crdsYaml)
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, r := range resources {
		result = append(result, r) // implizites cast zu runtime.Object
	}

	return result, nil
}
