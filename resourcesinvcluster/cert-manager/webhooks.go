package certmanager

import (
	"github.com/ScapeLanis/GoVCluster/structs"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func CreateWebhooksCertManager(namespace, clustername, version string) []runtime.Object {
	matchpolicy := admissionregistrationv1.MatchPolicyType("Equivalent")
	failurepolicy := admissionregistrationv1.FailurePolicyType("Fail")
	sideeffects := admissionregistrationv1.SideEffectClassNone
	cert_manager_webhook_mutating := &admissionregistrationv1.MutatingWebhookConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admissionregistration.k8s.io/v1",
			Kind:       "MutatingWebhookConfiguration",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-webhook",
			Labels: map[string]string{
				"app":                         "webhook",
				"app.kubernetes.io/name":      "webhook",
				"app.kubernetes.io/instance":  clustername,
				"app.kubernetes.io/component": "webhook",
				"app.kubernetes.io/version":   version,
			},
			Annotations: map[string]string{
				"cert-manager.io/inject-ca-from-secret": namespace + "/" + clustername + "-cert-manager-webhook-ca",
			},
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: "webhook.cert-manager.io",
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"cert-manager.io"},
							APIVersions: []string{"v1"},
							Resources:   []string{"certificaterequests"},
						},
						Operations: []admissionregistrationv1.OperationType{
							admissionregistrationv1.Create,
						},
					},
				},
				AdmissionReviewVersions: []string{"v1"},
				MatchPolicy:             &matchpolicy,
				TimeoutSeconds:          structs.Int32Ptr(30),
				FailurePolicy:           &failurepolicy,
				SideEffects:             &sideeffects,
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      clustername + "-cert-manager-webhook",
						Namespace: namespace,
						Path:      structs.StrPtr("/mutate"),
					},
				},
			},
		},
	}
	cert_manager_webhook_validating := &admissionregistrationv1.MutatingWebhookConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admissionregistration.k8s.io/v1",
			Kind:       "ValidatingWebhookConfiguration",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: clustername + "-cert-manager-webhook",
			Labels: map[string]string{
				"app":                         "webhook",
				"app.kubernetes.io/name":      "webhook",
				"app.kubernetes.io/instance":  clustername,
				"app.kubernetes.io/component": "webhook",
				"app.kubernetes.io/version":   version,
			},
			Annotations: map[string]string{
				"cert-manager.io/inject-ca-from-secret": namespace + "/" + clustername + "-cert-manager-webhook-ca",
			},
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: "webhook.cert-manager.io",
				NamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "cert-manager.io/disable-validation",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"true"},
						},
					},
				},
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"cert-manager.io", "acme.cert-manager.io"},
							APIVersions: []string{"v1"},
							Resources:   []string{"*/*"},
						},
						Operations: []admissionregistrationv1.OperationType{
							admissionregistrationv1.Create,
							admissionregistrationv1.Update,
						},
					},
				},
				AdmissionReviewVersions: []string{"v1"},
				MatchPolicy:             &matchpolicy,
				TimeoutSeconds:          structs.Int32Ptr(30),
				FailurePolicy:           &failurepolicy,
				SideEffects:             &sideeffects,
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      clustername + "-cert-manager-webhook",
						Namespace: namespace,
						Path:      structs.StrPtr("/validate"),
					},
				},
			},
		},
	}
	return []runtime.Object{
		cert_manager_webhook_mutating,
		cert_manager_webhook_validating,
	}
}
