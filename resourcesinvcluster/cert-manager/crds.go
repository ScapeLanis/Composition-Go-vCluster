package certmanager

import (
	"fmt"
	"io"
	"os"

	"github.com/crossplane/function-sdk-go/resource/composed"
	"gopkg.in/yaml.v3"
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

		// Leeres Dokument überspringen
		if obj == nil || len(obj) == 0 {
			continue
		}

		// Prüfe auf mandatory Felder
		if obj["apiVersion"] == nil || obj["kind"] == nil {
			// Ungültiges Kubernetes-Objekt, überspringen oder Fehler je nach Use-Case
			continue
		}

		c := composed.New()
		c.SetUnstructuredContent(obj)

		result = append(result, c)
	}

	return result, nil
}
func CreateCRDCertManager(namespace, clustername, version string) ([]*composed.Unstructured, error) {
	// Lade Ressourcen aus externer YAML-Datei
	resources, err := loadConfig("cert-manager-crds.yaml")
	if err != nil {
		return nil, err
	}
	return resources, nil
}
