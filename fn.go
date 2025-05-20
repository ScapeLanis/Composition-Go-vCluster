package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	sigsyaml "sigs.k8s.io/yaml"

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

	serviceaccount_vc := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vc-" + clustername,
			Namespace: namespace,
			Labels: map[string]string{
				"app":     "vcluster-" + clustername,
				"release": clustername,
			},
		},
	}
	serviceaccount_workload := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vc-workload-" + clustername,
			Namespace: namespace,
			Labels: map[string]string{
				"app":     "vcluster-" + clustername,
				"release": clustername,
			},
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vc-config-" + clustername,
			Namespace: namespace,
			Labels: map[string]string{
				"app":     "vcluster-" + clustername,
				"release": clustername,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"config.yaml": []byte("Y29udHJvbFBsYW5lOgogIGFkdmFuY2VkOgogICAgZGVmYXVsdEltYWdlUmVnaXN0cnk6ICIiCiAgICBnbG9iYWxNZXRhZGF0YToKICAgICAgYW5ub3RhdGlvbnM6IHt9CiAgICBoZWFkbGVzc1NlcnZpY2U6CiAgICAgIGFubm90YXRpb25zOiB7fQogICAgICBsYWJlbHM6IHt9CiAgICBzZXJ2aWNlQWNjb3VudDoKICAgICAgYW5ub3RhdGlvbnM6IHt9CiAgICAgIGVuYWJsZWQ6IHRydWUKICAgICAgaW1hZ2VQdWxsU2VjcmV0czogW10KICAgICAgbGFiZWxzOiB7fQogICAgICBuYW1lOiAiIgogICAgdmlydHVhbFNjaGVkdWxlcjoKICAgICAgZW5hYmxlZDogZmFsc2UKICAgIHdvcmtsb2FkU2VydmljZUFjY291bnQ6CiAgICAgIGFubm90YXRpb25zOiB7fQogICAgICBlbmFibGVkOiB0cnVlCiAgICAgIGltYWdlUHVsbFNlY3JldHM6IFtdCiAgICAgIGxhYmVsczoge30KICAgICAgbmFtZTogIiIKICBiYWNraW5nU3RvcmU6CiAgICBkYXRhYmFzZToKICAgICAgZW1iZWRkZWQ6CiAgICAgICAgZW5hYmxlZDogZmFsc2UKICAgICAgZXh0ZXJuYWw6CiAgICAgICAgY2FGaWxlOiAiIgogICAgICAgIGNlcnRGaWxlOiAiIgogICAgICAgIGNvbm5lY3RvcjogIiIKICAgICAgICBkYXRhU291cmNlOiAiIgogICAgICAgIGVuYWJsZWQ6IGZhbHNlCiAgICAgICAga2V5RmlsZTogIiIKICAgIGV0Y2Q6CiAgICAgIGRlcGxveToKICAgICAgICBlbmFibGVkOiBmYWxzZQogICAgICAgIGhlYWRsZXNzU2VydmljZToKICAgICAgICAgIGFubm90YXRpb25zOiB7fQogICAgICAgIHNlcnZpY2U6CiAgICAgICAgICBhbm5vdGF0aW9uczoge30KICAgICAgICAgIGVuYWJsZWQ6IHRydWUKICAgICAgICBzdGF0ZWZ1bFNldDoKICAgICAgICAgIGFubm90YXRpb25zOiB7fQogICAgICAgICAgZW5hYmxlU2VydmljZUxpbmtzOiB0cnVlCiAgICAgICAgICBlbmFibGVkOiB0cnVlCiAgICAgICAgICBlbnY6IFtdCiAgICAgICAgICBleHRyYUFyZ3M6IFtdCiAgICAgICAgICBoaWdoQXZhaWxhYmlsaXR5OgogICAgICAgICAgICByZXBsaWNhczogMQogICAgICAgICAgaW1hZ2U6CiAgICAgICAgICAgIHJlZ2lzdHJ5OiByZWdpc3RyeS5rOHMuaW8KICAgICAgICAgICAgcmVwb3NpdG9yeTogZXRjZAogICAgICAgICAgICB0YWc6IDMuNS4xNy0wCiAgICAgICAgICBpbWFnZVB1bGxQb2xpY3k6ICIiCiAgICAgICAgICBsYWJlbHM6IHt9CiAgICAgICAgICBwZXJzaXN0ZW5jZToKICAgICAgICAgICAgYWRkVm9sdW1lTW91bnRzOiBbXQogICAgICAgICAgICBhZGRWb2x1bWVzOiBbXQogICAgICAgICAgICB2b2x1bWVDbGFpbToKICAgICAgICAgICAgICBhY2Nlc3NNb2RlczoKICAgICAgICAgICAgICAtIFJlYWRXcml0ZU9uY2UKICAgICAgICAgICAgICBlbmFibGVkOiB0cnVlCiAgICAgICAgICAgICAgcmV0ZW50aW9uUG9saWN5OiBSZXRhaW4KICAgICAgICAgICAgICBzaXplOiA1R2kKICAgICAgICAgICAgICBzdG9yYWdlQ2xhc3M6ICIiCiAgICAgICAgICAgIHZvbHVtZUNsYWltVGVtcGxhdGVzOiBbXQogICAgICAgICAgcG9kczoKICAgICAgICAgICAgYW5ub3RhdGlvbnM6IHt9CiAgICAgICAgICAgIGxhYmVsczoge30KICAgICAgICAgIHJlc291cmNlczoKICAgICAgICAgICAgcmVxdWVzdHM6CiAgICAgICAgICAgICAgY3B1OiAyMG0KICAgICAgICAgICAgICBtZW1vcnk6IDE1ME1pCiAgICAgICAgICBzY2hlZHVsaW5nOgogICAgICAgICAgICBhZmZpbml0eToge30KICAgICAgICAgICAgbm9kZVNlbGVjdG9yOiB7fQogICAgICAgICAgICBwb2RNYW5hZ2VtZW50UG9saWN5OiBQYXJhbGxlbAogICAgICAgICAgICBwcmlvcml0eUNsYXNzTmFtZTogIiIKICAgICAgICAgICAgdG9sZXJhdGlvbnM6IFtdCiAgICAgICAgICAgIHRvcG9sb2d5U3ByZWFkQ29uc3RyYWludHM6IFtdCiAgICAgICAgICBzZWN1cml0eToKICAgICAgICAgICAgY29udGFpbmVyU2VjdXJpdHlDb250ZXh0OiB7fQogICAgICAgICAgICBwb2RTZWN1cml0eUNvbnRleHQ6IHt9CiAgICAgIGVtYmVkZGVkOgogICAgICAgIGVuYWJsZWQ6IGZhbHNlCiAgICAgICAgbWlncmF0ZUZyb21EZXBsb3llZEV0Y2Q6IGZhbHNlCiAgY29yZWRuczoKICAgIGRlcGxveW1lbnQ6CiAgICAgIGFmZmluaXR5OiB7fQogICAgICBhbm5vdGF0aW9uczoge30KICAgICAgaW1hZ2U6ICIiCiAgICAgIGxhYmVsczoge30KICAgICAgbm9kZVNlbGVjdG9yOiB7fQogICAgICBwb2RzOgogICAgICAgIGFubm90YXRpb25zOiB7fQogICAgICAgIGxhYmVsczoge30KICAgICAgcmVwbGljYXM6IDEKICAgICAgcmVzb3VyY2VzOgogICAgICAgIGxpbWl0czoKICAgICAgICAgIGNwdTogMTAwMG0KICAgICAgICAgIG1lbW9yeTogMTcwTWkKICAgICAgICByZXF1ZXN0czoKICAgICAgICAgIGNwdTogMjBtCiAgICAgICAgICBtZW1vcnk6IDY0TWkKICAgICAgdG9sZXJhdGlvbnM6IFtdCiAgICAgIHRvcG9sb2d5U3ByZWFkQ29uc3RyYWludHM6CiAgICAgIC0gbGFiZWxTZWxlY3RvcjoKICAgICAgICAgIG1hdGNoTGFiZWxzOgogICAgICAgICAgICBrOHMtYXBwOiB2Y2x1c3Rlci1rdWJlLWRucwogICAgICAgIG1heFNrZXc6IDEKICAgICAgICB0b3BvbG9neUtleToga3ViZXJuZXRlcy5pby9ob3N0bmFtZQogICAgICAgIHdoZW5VbnNhdGlzZmlhYmxlOiBEb05vdFNjaGVkdWxlCiAgICBlbWJlZGRlZDogZmFsc2UKICAgIGVuYWJsZWQ6IHRydWUKICAgIG92ZXJ3cml0ZUNvbmZpZzogIiIKICAgIG92ZXJ3cml0ZU1hbmlmZXN0czogIiIKICAgIHByaW9yaXR5Q2xhc3NOYW1lOiAiIgogICAgc2VydmljZToKICAgICAgYW5ub3RhdGlvbnM6IHt9CiAgICAgIGxhYmVsczoge30KICAgICAgc3BlYzoKICAgICAgICB0eXBlOiBDbHVzdGVySVAKICBkaXN0cm86CiAgICBrMHM6CiAgICAgIGNvbW1hbmQ6IFtdCiAgICAgIGNvbmZpZzogIiIKICAgICAgZW5hYmxlZDogZmFsc2UKICAgICAgZXh0cmFBcmdzOiBbXQogICAgICBpbWFnZToKICAgICAgICByZWdpc3RyeTogIiIKICAgICAgICByZXBvc2l0b3J5OiBrMHNwcm9qZWN0L2swcwogICAgICAgIHRhZzogdjEuMzAuMi1rMHMuMAogICAgICBpbWFnZVB1bGxQb2xpY3k6ICIiCiAgICAgIHJlc291cmNlczoKICAgICAgICBsaW1pdHM6CiAgICAgICAgICBjcHU6IDEwMG0KICAgICAgICAgIG1lbW9yeTogMjU2TWkKICAgICAgICByZXF1ZXN0czoKICAgICAgICAgIGNwdTogNDBtCiAgICAgICAgICBtZW1vcnk6IDY0TWkKICAgICAgc2VjdXJpdHlDb250ZXh0OiB7fQogICAgazNzOgogICAgICBjb21tYW5kOiBbXQogICAgICBlbmFibGVkOiBmYWxzZQogICAgICBleHRyYUFyZ3M6IFtdCiAgICAgIGltYWdlOgogICAgICAgIHJlZ2lzdHJ5OiAiIgogICAgICAgIHJlcG9zaXRvcnk6IHJhbmNoZXIvazNzCiAgICAgICAgdGFnOiB2MS4zMi4xLWszczEKICAgICAgaW1hZ2VQdWxsUG9saWN5OiAiIgogICAgICByZXNvdXJjZXM6CiAgICAgICAgbGltaXRzOgogICAgICAgICAgY3B1OiAxMDBtCiAgICAgICAgICBtZW1vcnk6IDI1Nk1pCiAgICAgICAgcmVxdWVzdHM6CiAgICAgICAgICBjcHU6IDQwbQogICAgICAgICAgbWVtb3J5OiA2NE1pCiAgICAgIHNlY3VyaXR5Q29udGV4dDoge30KICAgIGs4czoKICAgICAgYXBpU2VydmVyOgogICAgICAgIGNvbW1hbmQ6IFtdCiAgICAgICAgZW5hYmxlZDogdHJ1ZQogICAgICAgIGV4dHJhQXJnczogW10KICAgICAgICBpbWFnZToKICAgICAgICAgIHJlZ2lzdHJ5OiByZWdpc3RyeS5rOHMuaW8KICAgICAgICAgIHJlcG9zaXRvcnk6IGt1YmUtYXBpc2VydmVyCiAgICAgICAgICB0YWc6IHYxLjMyLjEKICAgICAgICBpbWFnZVB1bGxQb2xpY3k6ICIiCiAgICAgIGNvbnRyb2xsZXJNYW5hZ2VyOgogICAgICAgIGNvbW1hbmQ6IFtdCiAgICAgICAgZW5hYmxlZDogdHJ1ZQogICAgICAgIGV4dHJhQXJnczogW10KICAgICAgICBpbWFnZToKICAgICAgICAgIHJlZ2lzdHJ5OiByZWdpc3RyeS5rOHMuaW8KICAgICAgICAgIHJlcG9zaXRvcnk6IGt1YmUtY29udHJvbGxlci1tYW5hZ2VyCiAgICAgICAgICB0YWc6IHYxLjMyLjEKICAgICAgICBpbWFnZVB1bGxQb2xpY3k6ICIiCiAgICAgIGVuYWJsZWQ6IGZhbHNlCiAgICAgIGVudjogW10KICAgICAgcmVzb3VyY2VzOgogICAgICAgIGxpbWl0czoKICAgICAgICAgIGNwdTogMTAwbQogICAgICAgICAgbWVtb3J5OiAyNTZNaQogICAgICAgIHJlcXVlc3RzOgogICAgICAgICAgY3B1OiA0MG0KICAgICAgICAgIG1lbW9yeTogNjRNaQogICAgICBzY2hlZHVsZXI6CiAgICAgICAgY29tbWFuZDogW10KICAgICAgICBleHRyYUFyZ3M6IFtdCiAgICAgICAgaW1hZ2U6CiAgICAgICAgICByZWdpc3RyeTogcmVnaXN0cnkuazhzLmlvCiAgICAgICAgICByZXBvc2l0b3J5OiBrdWJlLXNjaGVkdWxlcgogICAgICAgICAgdGFnOiB2MS4zMi4xCiAgICAgICAgaW1hZ2VQdWxsUG9saWN5OiAiIgogICAgICBzZWN1cml0eUNvbnRleHQ6IHt9CiAgICAgIHZlcnNpb246ICIiCiAgaW5ncmVzczoKICAgIGFubm90YXRpb25zOgogICAgICBuZ2lueC5pbmdyZXNzLmt1YmVybmV0ZXMuaW8vYmFja2VuZC1wcm90b2NvbDogSFRUUFMKICAgICAgbmdpbnguaW5ncmVzcy5rdWJlcm5ldGVzLmlvL3NzbC1wYXNzdGhyb3VnaDogInRydWUiCiAgICAgIG5naW54LmluZ3Jlc3Mua3ViZXJuZXRlcy5pby9zc2wtcmVkaXJlY3Q6ICJ0cnVlIgogICAgZW5hYmxlZDogZmFsc2UKICAgIGhvc3Q6IG15LWhvc3QuY29tCiAgICBsYWJlbHM6IHt9CiAgICBwYXRoVHlwZTogSW1wbGVtZW50YXRpb25TcGVjaWZpYwogICAgc3BlYzoKICAgICAgdGxzOiBbXQogIHByb3h5OgogICAgYmluZEFkZHJlc3M6IDAuMC4wLjAKICAgIGV4dHJhU0FOczogW10KICAgIHBvcnQ6IDg0NDMKICBzZXJ2aWNlOgogICAgYW5ub3RhdGlvbnM6IHt9CiAgICBlbmFibGVkOiB0cnVlCiAgICBodHRwc05vZGVQb3J0OiAwCiAgICBrdWJlbGV0Tm9kZVBvcnQ6IDAKICAgIGxhYmVsczoge30KICAgIHNwZWM6CiAgICAgIHR5cGU6IENsdXN0ZXJJUAogIHNlcnZpY2VNb25pdG9yOgogICAgYW5ub3RhdGlvbnM6IHt9CiAgICBlbmFibGVkOiBmYWxzZQogICAgbGFiZWxzOiB7fQogIHN0YXRlZnVsU2V0OgogICAgYW5ub3RhdGlvbnM6IHt9CiAgICBhcmdzOiBbXQogICAgY29tbWFuZDogW10KICAgIGVuYWJsZVNlcnZpY2VMaW5rczogdHJ1ZQogICAgZW52OiBbXQogICAgaGlnaEF2YWlsYWJpbGl0eToKICAgICAgbGVhc2VEdXJhdGlvbjogNjAKICAgICAgcmVuZXdEZWFkbGluZTogNDAKICAgICAgcmVwbGljYXM6IDEKICAgICAgcmV0cnlQZXJpb2Q6IDE1CiAgICBpbWFnZToKICAgICAgcmVnaXN0cnk6IGdoY3IuaW8KICAgICAgcmVwb3NpdG9yeTogbG9mdC1zaC92Y2x1c3Rlci1wcm8KICAgICAgdGFnOiAiIgogICAgaW1hZ2VQdWxsUG9saWN5OiAiIgogICAgbGFiZWxzOiB7fQogICAgcGVyc2lzdGVuY2U6CiAgICAgIGFkZFZvbHVtZU1vdW50czogW10KICAgICAgYWRkVm9sdW1lczogW10KICAgICAgYmluYXJpZXNWb2x1bWU6CiAgICAgIC0gZW1wdHlEaXI6IHt9CiAgICAgICAgbmFtZTogYmluYXJpZXMKICAgICAgZGF0YVZvbHVtZTogW10KICAgICAgdm9sdW1lQ2xhaW06CiAgICAgICAgYWNjZXNzTW9kZXM6CiAgICAgICAgLSBSZWFkV3JpdGVPbmNlCiAgICAgICAgZW5hYmxlZDogYXV0bwogICAgICAgIHJldGVudGlvblBvbGljeTogUmV0YWluCiAgICAgICAgc2l6ZTogNUdpCiAgICAgICAgc3RvcmFnZUNsYXNzOiAiIgogICAgICB2b2x1bWVDbGFpbVRlbXBsYXRlczogW10KICAgIHBvZHM6CiAgICAgIGFubm90YXRpb25zOiB7fQogICAgICBsYWJlbHM6IHt9CiAgICBwcm9iZXM6CiAgICAgIGxpdmVuZXNzUHJvYmU6CiAgICAgICAgZW5hYmxlZDogdHJ1ZQogICAgICByZWFkaW5lc3NQcm9iZToKICAgICAgICBlbmFibGVkOiB0cnVlCiAgICAgIHN0YXJ0dXBQcm9iZToKICAgICAgICBlbmFibGVkOiB0cnVlCiAgICByZXNvdXJjZXM6CiAgICAgIGxpbWl0czoKICAgICAgICBlcGhlbWVyYWwtc3RvcmFnZTogOEdpCiAgICAgICAgbWVtb3J5OiAyR2kKICAgICAgcmVxdWVzdHM6CiAgICAgICAgY3B1OiAyMDBtCiAgICAgICAgZXBoZW1lcmFsLXN0b3JhZ2U6IDQwME1pCiAgICAgICAgbWVtb3J5OiAyNTZNaQogICAgc2NoZWR1bGluZzoKICAgICAgYWZmaW5pdHk6IHt9CiAgICAgIG5vZGVTZWxlY3Rvcjoge30KICAgICAgcG9kTWFuYWdlbWVudFBvbGljeTogUGFyYWxsZWwKICAgICAgcHJpb3JpdHlDbGFzc05hbWU6ICIiCiAgICAgIHRvbGVyYXRpb25zOiBbXQogICAgICB0b3BvbG9neVNwcmVhZENvbnN0cmFpbnRzOiBbXQogICAgc2VjdXJpdHk6CiAgICAgIGNvbnRhaW5lclNlY3VyaXR5Q29udGV4dDoKICAgICAgICBhbGxvd1ByaXZpbGVnZUVzY2FsYXRpb246IGZhbHNlCiAgICAgICAgcnVuQXNHcm91cDogMAogICAgICAgIHJ1bkFzVXNlcjogMAogICAgICBwb2RTZWN1cml0eUNvbnRleHQ6IHt9CiAgICB3b3JraW5nRGlyOiAiIgpleHBlcmltZW50YWw6CiAgZGVwbG95OgogICAgaG9zdDoKICAgICAgbWFuaWZlc3RzOiAiIgogICAgICBtYW5pZmVzdHNUZW1wbGF0ZTogIiIKICAgIHZjbHVzdGVyOgogICAgICBoZWxtOiBbXQogICAgICBtYW5pZmVzdHM6ICIiCiAgICAgIG1hbmlmZXN0c1RlbXBsYXRlOiAiIgogIGdlbmVyaWNTeW5jOgogICAgY2x1c3RlclJvbGU6CiAgICAgIGV4dHJhUnVsZXM6IFtdCiAgICByb2xlOgogICAgICBleHRyYVJ1bGVzOiBbXQogIGlzb2xhdGVkQ29udHJvbFBsYW5lOgogICAgaGVhZGxlc3M6IGZhbHNlCiAgbXVsdGlOYW1lc3BhY2VNb2RlOgogICAgZW5hYmxlZDogZmFsc2UKICByZXVzZU5hbWVzcGFjZTogZmFsc2UKICBzeW5jU2V0dGluZ3M6CiAgICBkaXNhYmxlU3luYzogZmFsc2UKICAgIHJld3JpdGVLdWJlcm5ldGVzU2VydmljZTogZmFsc2UKICAgIHNldE93bmVyOiB0cnVlCiAgICB0YXJnZXROYW1lc3BhY2U6ICIiCmV4cG9ydEt1YmVDb25maWc6CiAgY29udGV4dDogIiIKICBpbnNlY3VyZTogZmFsc2UKICBzZWNyZXQ6CiAgICBuYW1lOiAiIgogICAgbmFtZXNwYWNlOiAiIgogIHNlcnZlcjogIiIKICBzZXJ2aWNlQWNjb3VudDoKICAgIGNsdXN0ZXJSb2xlOiAiIgogICAgbmFtZTogIiIKICAgIG5hbWVzcGFjZTogIiIKZXh0ZXJuYWw6IHt9CmludGVncmF0aW9uczoKICBjZXJ0TWFuYWdlcjoKICAgIGVuYWJsZWQ6IGZhbHNlCiAgICBzeW5jOgogICAgICBmcm9tSG9zdDoKICAgICAgICBjbHVzdGVySXNzdWVyczoKICAgICAgICAgIGVuYWJsZWQ6IHRydWUKICAgICAgICAgIHNlbGVjdG9yOgogICAgICAgICAgICBsYWJlbHM6IHt9CiAgICAgIHRvSG9zdDoKICAgICAgICBjZXJ0aWZpY2F0ZXM6CiAgICAgICAgICBlbmFibGVkOiB0cnVlCiAgICAgICAgaXNzdWVyczoKICAgICAgICAgIGVuYWJsZWQ6IHRydWUKICBleHRlcm5hbFNlY3JldHM6CiAgICBlbmFibGVkOiBmYWxzZQogICAgc3luYzoKICAgICAgY2x1c3RlclN0b3JlczoKICAgICAgICBlbmFibGVkOiBmYWxzZQogICAgICAgIHNlbGVjdG9yOgogICAgICAgICAgbGFiZWxzOiB7fQogICAgICBleHRlcm5hbFNlY3JldHM6CiAgICAgICAgZW5hYmxlZDogdHJ1ZQogICAgICBzdG9yZXM6CiAgICAgICAgZW5hYmxlZDogZmFsc2UKICAgIHdlYmhvb2s6CiAgICAgIGVuYWJsZWQ6IGZhbHNlCiAga3ViZVZpcnQ6CiAgICBlbmFibGVkOiBmYWxzZQogICAgc3luYzoKICAgICAgZGF0YVZvbHVtZXM6CiAgICAgICAgZW5hYmxlZDogZmFsc2UKICAgICAgdmlydHVhbE1hY2hpbmVDbG9uZXM6CiAgICAgICAgZW5hYmxlZDogdHJ1ZQogICAgICB2aXJ0dWFsTWFjaGluZUluc3RhbmNlTWlncmF0aW9uczoKICAgICAgICBlbmFibGVkOiB0cnVlCiAgICAgIHZpcnR1YWxNYWNoaW5lSW5zdGFuY2VzOgogICAgICAgIGVuYWJsZWQ6IHRydWUKICAgICAgdmlydHVhbE1hY2hpbmVQb29sczoKICAgICAgICBlbmFibGVkOiB0cnVlCiAgICAgIHZpcnR1YWxNYWNoaW5lczoKICAgICAgICBlbmFibGVkOiB0cnVlCiAgICB3ZWJob29rOgogICAgICBlbmFibGVkOiB0cnVlCiAgbWV0cmljc1NlcnZlcjoKICAgIGVuYWJsZWQ6IGZhbHNlCiAgICBub2RlczogdHJ1ZQogICAgcG9kczogdHJ1ZQpuZXR3b3JraW5nOgogIGFkdmFuY2VkOgogICAgY2x1c3RlckRvbWFpbjogY2x1c3Rlci5sb2NhbAogICAgZmFsbGJhY2tIb3N0Q2x1c3RlcjogZmFsc2UKICAgIHByb3h5S3ViZWxldHM6CiAgICAgIGJ5SG9zdG5hbWU6IHRydWUKICAgICAgYnlJUDogdHJ1ZQogIHJlcGxpY2F0ZVNlcnZpY2VzOgogICAgZnJvbUhvc3Q6IFtdCiAgICB0b0hvc3Q6IFtdCiAgcmVzb2x2ZUROUzogW10KcGx1Z2luczoge30KcG9saWNpZXM6CiAgY2VudHJhbEFkbWlzc2lvbjoKICAgIG11dGF0aW5nV2ViaG9va3M6IFtdCiAgICB2YWxpZGF0aW5nV2ViaG9va3M6IFtdCiAgbGltaXRSYW5nZToKICAgIGFubm90YXRpb25zOiB7fQogICAgZGVmYXVsdDoKICAgICAgY3B1OiAiMSIKICAgICAgZXBoZW1lcmFsLXN0b3JhZ2U6IDhHaQogICAgICBtZW1vcnk6IDUxMk1pCiAgICBkZWZhdWx0UmVxdWVzdDoKICAgICAgY3B1OiAxMDBtCiAgICAgIGVwaGVtZXJhbC1zdG9yYWdlOiAzR2kKICAgICAgbWVtb3J5OiAxMjhNaQogICAgZW5hYmxlZDogYXV0bwogICAgbGFiZWxzOiB7fQogICAgbWF4OiB7fQogICAgbWluOiB7fQogIG5ldHdvcmtQb2xpY3k6CiAgICBhbm5vdGF0aW9uczoge30KICAgIGVuYWJsZWQ6IGZhbHNlCiAgICBmYWxsYmFja0RuczogOC44LjguOAogICAgbGFiZWxzOiB7fQogICAgb3V0Z29pbmdDb25uZWN0aW9uczoKICAgICAgaXBCbG9jazoKICAgICAgICBjaWRyOiAwLjAuMC4wLzAKICAgICAgICBleGNlcHQ6CiAgICAgICAgLSAxMDAuNjQuMC4wLzEwCiAgICAgICAgLSAxMjcuMC4wLjAvOAogICAgICAgIC0gMTAuMC4wLjAvOAogICAgICAgIC0gMTcyLjE2LjAuMC8xMgogICAgICAgIC0gMTkyLjE2OC4wLjAvMTYKICAgICAgcGxhdGZvcm06IHRydWUKICByZXNvdXJjZVF1b3RhOgogICAgYW5ub3RhdGlvbnM6IHt9CiAgICBlbmFibGVkOiBhdXRvCiAgICBsYWJlbHM6IHt9CiAgICBxdW90YToKICAgICAgY291bnQvY29uZmlnbWFwczogMTAwCiAgICAgIGNvdW50L2VuZHBvaW50czogNDAKICAgICAgY291bnQvcGVyc2lzdGVudHZvbHVtZWNsYWltczogMjAKICAgICAgY291bnQvcG9kczogMjAKICAgICAgY291bnQvc2VjcmV0czogMTAwCiAgICAgIGNvdW50L3NlcnZpY2VzOiAyMAogICAgICBsaW1pdHMuY3B1OiAyMAogICAgICBsaW1pdHMuZXBoZW1lcmFsLXN0b3JhZ2U6IDE2MEdpCiAgICAgIGxpbWl0cy5tZW1vcnk6IDQwR2kKICAgICAgcmVxdWVzdHMuY3B1OiAxMAogICAgICByZXF1ZXN0cy5lcGhlbWVyYWwtc3RvcmFnZTogNjBHaQogICAgICByZXF1ZXN0cy5tZW1vcnk6IDIwR2kKICAgICAgcmVxdWVzdHMuc3RvcmFnZTogMTAwR2kKICAgICAgc2VydmljZXMubG9hZGJhbGFuY2VyczogMQogICAgICBzZXJ2aWNlcy5ub2RlcG9ydHM6IDAKICAgIHNjb3BlU2VsZWN0b3I6CiAgICAgIG1hdGNoRXhwcmVzc2lvbnM6IFtdCiAgICBzY29wZXM6IFtdCnJiYWM6CiAgY2x1c3RlclJvbGU6CiAgICBlbmFibGVkOiBhdXRvCiAgICBleHRyYVJ1bGVzOiBbXQogICAgb3ZlcndyaXRlUnVsZXM6IFtdCiAgcm9sZToKICAgIGVuYWJsZWQ6IHRydWUKICAgIGV4dHJhUnVsZXM6IFtdCiAgICBvdmVyd3JpdGVSdWxlczogW10Kc3luYzoKICBmcm9tSG9zdDoKICAgIGNvbmZpZ01hcHM6CiAgICAgIGVuYWJsZWQ6IGZhbHNlCiAgICAgIG1hcHBpbmdzOgogICAgICAgIGJ5TmFtZToge30KICAgIGNzaURyaXZlcnM6CiAgICAgIGVuYWJsZWQ6IGF1dG8KICAgIGNzaU5vZGVzOgogICAgICBlbmFibGVkOiBhdXRvCiAgICBjc2lTdG9yYWdlQ2FwYWNpdGllczoKICAgICAgZW5hYmxlZDogYXV0bwogICAgZXZlbnRzOgogICAgICBlbmFibGVkOiB0cnVlCiAgICBpbmdyZXNzQ2xhc3NlczoKICAgICAgZW5hYmxlZDogZmFsc2UKICAgIG5vZGVzOgogICAgICBjbGVhckltYWdlU3RhdHVzOiBmYWxzZQogICAgICBlbmFibGVkOiBmYWxzZQogICAgICBzZWxlY3RvcjoKICAgICAgICBhbGw6IGZhbHNlCiAgICAgICAgbGFiZWxzOiB7fQogICAgICBzeW5jQmFja0NoYW5nZXM6IGZhbHNlCiAgICBwcmlvcml0eUNsYXNzZXM6CiAgICAgIGVuYWJsZWQ6IGZhbHNlCiAgICBydW50aW1lQ2xhc3NlczoKICAgICAgZW5hYmxlZDogZmFsc2UKICAgIHNlY3JldHM6CiAgICAgIGVuYWJsZWQ6IGZhbHNlCiAgICAgIG1hcHBpbmdzOgogICAgICAgIGJ5TmFtZToge30KICAgIHN0b3JhZ2VDbGFzc2VzOgogICAgICBlbmFibGVkOiBhdXRvCiAgICB2b2x1bWVTbmFwc2hvdENsYXNzZXM6CiAgICAgIGVuYWJsZWQ6IGZhbHNlCiAgdG9Ib3N0OgogICAgY29uZmlnTWFwczoKICAgICAgYWxsOiBmYWxzZQogICAgICBlbmFibGVkOiB0cnVlCiAgICBlbmRwb2ludHM6CiAgICAgIGVuYWJsZWQ6IHRydWUKICAgIGluZ3Jlc3NlczoKICAgICAgZW5hYmxlZDogZmFsc2UKICAgIG5ldHdvcmtQb2xpY2llczoKICAgICAgZW5hYmxlZDogZmFsc2UKICAgIHBlcnNpc3RlbnRWb2x1bWVDbGFpbXM6CiAgICAgIGVuYWJsZWQ6IHRydWUKICAgIHBlcnNpc3RlbnRWb2x1bWVzOgogICAgICBlbmFibGVkOiBmYWxzZQogICAgcG9kRGlzcnVwdGlvbkJ1ZGdldHM6CiAgICAgIGVuYWJsZWQ6IGZhbHNlCiAgICBwb2RzOgogICAgICBlbmFibGVkOiB0cnVlCiAgICAgIGVuZm9yY2VUb2xlcmF0aW9uczogW10KICAgICAgcHJpb3JpdHlDbGFzc05hbWU6ICIiCiAgICAgIHJld3JpdGVIb3N0czoKICAgICAgICBlbmFibGVkOiB0cnVlCiAgICAgICAgaW5pdENvbnRhaW5lcjoKICAgICAgICAgIGltYWdlOiBsaWJyYXJ5L2FscGluZTozLjIwCiAgICAgICAgICByZXNvdXJjZXM6CiAgICAgICAgICAgIGxpbWl0czoKICAgICAgICAgICAgICBjcHU6IDMwbQogICAgICAgICAgICAgIG1lbW9yeTogNjRNaQogICAgICAgICAgICByZXF1ZXN0czoKICAgICAgICAgICAgICBjcHU6IDMwbQogICAgICAgICAgICAgIG1lbW9yeTogNjRNaQogICAgICBydW50aW1lQ2xhc3NOYW1lOiAiIgogICAgICB0cmFuc2xhdGVJbWFnZToge30KICAgICAgdXNlU2VjcmV0c0ZvclNBVG9rZW5zOiBmYWxzZQogICAgcHJpb3JpdHlDbGFzc2VzOgogICAgICBlbmFibGVkOiBmYWxzZQogICAgc2VjcmV0czoKICAgICAgYWxsOiBmYWxzZQogICAgICBlbmFibGVkOiB0cnVlCiAgICBzZXJ2aWNlQWNjb3VudHM6CiAgICAgIGVuYWJsZWQ6IGZhbHNlCiAgICBzZXJ2aWNlczoKICAgICAgZW5hYmxlZDogdHJ1ZQogICAgc3RvcmFnZUNsYXNzZXM6CiAgICAgIGVuYWJsZWQ6IGZhbHNlCiAgICB2b2x1bWVTbmFwc2hvdENvbnRlbnRzOgogICAgICBlbmFibGVkOiBmYWxzZQogICAgdm9sdW1lU25hcHNob3RzOgogICAgICBlbmFibGVkOiBmYWxzZQp0ZWxlbWV0cnk6CiAgZW5hYmxlZDogdHJ1ZQ=="),
		},
	}
	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vc-coredns-" + clustername,
			Namespace: namespace,
		},
		Data: map[string]string{
			"coredns.yaml": `apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: coredns
      namespace: kube-system
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      labels:
        kubernetes.io/bootstrapping: rbac-defaults
      name: system:coredns
    rules:
      - apiGroups:
          - ""
        resources:
          - endpoints
          - services
          - pods
          - namespaces
        verbs:
          - list
          - watch
      - apiGroups:
          - discovery.k8s.io
        resources:
          - endpointslices
        verbs:
          - list
          - watch
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      annotations:
        rbac.authorization.kubernetes.io/autoupdate: "true"
      labels:
        kubernetes.io/bootstrapping: rbac-defaults
      name: system:coredns
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: system:coredns
    subjects:
      - kind: ServiceAccount
        name: coredns
        namespace: kube-system
    ---
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: coredns
      namespace: kube-system
    data:
      Corefile: |-
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
      
        import /etc/coredns/custom/*.server
      NodeHosts: ""
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: coredns
      namespace: kube-system
      labels:
        k8s-app: vcluster-kube-dns
        kubernetes.io/name: "CoreDNS"
    spec:
      replicas: 1
      strategy:
        type: RollingUpdate
        rollingUpdate:
          maxUnavailable: 1
      selector:
        matchLabels:
          k8s-app: vcluster-kube-dns
      template:
        metadata:
          labels:
            k8s-app: vcluster-kube-dns
        spec:
          priorityClassName: ""
          serviceAccountName: coredns
          nodeSelector:
            kubernetes.io/os: linux
          topologySpreadConstraints:
            - labelSelector:
                matchLabels:
                  k8s-app: vcluster-kube-dns
              maxSkew: 1
              topologyKey: kubernetes.io/hostname
              whenUnsatisfiable: DoNotSchedule
          containers:
            - name: coredns
              image: {{.IMAGE}}
              imagePullPolicy: IfNotPresent
              resources:
                limits:
                  cpu: 1000m
                  memory: 170Mi
                requests:
                  cpu: 20m
                  memory: 64Mi
              args: [ "-conf", "/etc/coredns/Corefile" ]
              volumeMounts:
                - name: config-volume
                  mountPath: /etc/coredns
                  readOnly: true
                - name: custom-config-volume
                  mountPath: /etc/coredns/custom
                  readOnly: true
              securityContext:
                runAsNonRoot: true
                runAsUser: {{.RUN_AS_USER}}
                runAsGroup: {{.RUN_AS_GROUP}}
                allowPrivilegeEscalation: false
                capabilities:
                  add:
                    - NET_BIND_SERVICE
                  drop:
                    - ALL
                readOnlyRootFilesystem: true
              livenessProbe:
                httpGet:
                  path: /health
                  port: 8080
                  scheme: HTTP
                initialDelaySeconds: 60
                periodSeconds: 10
                timeoutSeconds: 1
                successThreshold: 1
                failureThreshold: 3
              readinessProbe:
                httpGet:
                  path: /ready
                  port: 8181
                  scheme: HTTP
                initialDelaySeconds: 0
                periodSeconds: 2
                timeoutSeconds: 1
                successThreshold: 1
                failureThreshold: 3
          dnsPolicy: Default
          volumes:
            - name: config-volume
              configMap:
                name: coredns
                items:
                  - key: Corefile
                    path: Corefile
                  - key: NodeHosts
                    path: NodeHosts
            - name: custom-config-volume
              configMap:
                name: coredns-custom
                optional: true
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: kube-dns
      namespace: kube-system
      annotations:
        prometheus.io/port: "9153"
        prometheus.io/scrape: "true"
      labels:
        k8s-app: vcluster-kube-dns
        kubernetes.io/cluster-service: "true"
        kubernetes.io/name: "CoreDNS"
    spec:
      type: ClusterIP
      selector:
        k8s-app: vcluster-kube-dns
      ports:
        - name: dns
          port: 53
          targetPort: 1053
          protocol: UDP
        - name: dns-tcp
          port: 53
          targetPort: 1053
          protocol: TCP
        - name: metrics
          port: 9153
          protocol: TCP`,
		},
	}

	role := rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vc-" + clustername,
			Namespace: namespace,
			Labels: map[string]string{
				"app":     "vcluster-" + clustername,
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
	rolebinding := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vc-" + clustername,
			Namespace: namespace,
			Labels: map[string]string{
				"app":     "vcluster-" + clustername,
				"release": clustername,
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "vc-" + clustername,
				Namespace: clustername,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "vc-" + clustername,
		},
	}

	service_headless := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername + "-headless",
			Namespace: namespace,
			Labels: map[string]string{
				"app":     "vcluster-" + clustername,
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
				"app":     "vcluster-" + clustername,
				"release": clustername,
			},
		},
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clustername,
			Namespace: namespace,
			Labels: map[string]string{
				"app":                      "vcluster-" + clustername,
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
				"app":     "vcluster-" + clustername,
				"release": clustername,
			},
		},
	}

	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "",
			Namespace: "",
			Labels: map[string]string{
				"app": "",
				"release": "",
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "vcluster-" + clustername,
					"release": clustername,
				},
			},
			PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: appsv1.RetainPersistentVolumeClaimRetentionPolicyType,
			},
			ServiceName: clustername + "-headless",
			PodManagementPolicy: appsv1.ParallelPodManagement,
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						}
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: "5Gi"
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
					"app": "vcluster-" + clustername,
					"release": clustername
				},
			},
			Spec: corev1.PodSpec{
				TerminationGracePeriodSeconds: 10,
				ServiceAccountName: "vc-" + clustername,
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
								Name: "vc-coredns-" + clustername
							},
						},
					},
				},
				InitContainers: []corev1.Container{
					{
						Name: "vcluster-copy",
						Image: "ghcr.io/loft-sh/vcluster-pro:0.24.1",
						VolumeMounts: []corev1.VolumeMount{
							{
								MountPath: "/binaries",
								Name: "binaries",
							},
						},
						Command: []string{"/bin/sh"},
						Args: []string{"-c", "cp /vcluster /binaries/vcluster"},
						SecurityContext: &corev1.SecurityContext{},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceLimitsCPU: "100m",
								corev1.ResourceLimitsMemory: "256Mi",
							},
							Requests: corev1.ResourceList{
								corev1.ResourceRequestsCPU: "40m",
								corev1.ResourceRequestsMemory: "64Mi",
							},
						},
					},
					{
						Name: "kube-controller-manager"
						Image: "registry.k8s.io/kube-controller-manager:v1.32.0",
						VolumeMounts: []corev1.VolumeMount{
							{
								MountPath: "/binaries",
								Name: "binaries",
							}
						},
						Command: []string{"/binaries/vcluster"}
						Args: []string{"cp", "/usr/local/bin/kube-controller-manager", "/binaries/kube-controller-manager"},
						SecurityContext: &corev1.SecurityContext{},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceLimitsCPU: "100m",
								corev1.ResourceLimitsMemory: "256Mi"
							},
							Requests: corev1.ResourceList{
								corev1.ResourceRequestsCPU: "40m"
								corev1.ResourceRequestsMemory: "64Mi"
							},
						},
					},
					{
						Name: "kube-apiserver",
						Image: "registry.k8s.io/kube-apiserver:v1.32.0",
						VolumeMounts: []corev1.VolumeMount{
							{
								MountPath: "/binaries",
								Name: "binaries",
							}
						},
						Command: []string{"/binaries/vcluster"},
						Args: []string{"cp", "/usr/local/bin/kube-apiserver", "/binaries/kube-apiserver"},
						SecurityContext: &corev1.SecurityContext{},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceLimitsCPU: "100m"
								corev1.ResourceLimitsMemory: "256Mi"
							},
							Requests: corev1.ResourceList{
								corev1.ResourceRequestsCPU: "40m"
								corev1.ResourceRequestsMemory: "64Mi"
							},
						},
					}
				},
				EnableServiceLinks: true,
			},
		},
	}
},
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
			if err := setConnectionDetails(u, clustername, namespace); err != nil {
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

// gibt eine methode in composed mit setConnectionDetails
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
