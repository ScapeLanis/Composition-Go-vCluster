package structs

import (
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Hilfsfunktionen
func Int32Ptr(i int32) *int32 { return &i }
func Int64Ptr(i int64) *int64 { return &i }
func BoolPtr(b bool) *bool    { return &b }
func StrPtr(s string) *string { return &s }
func IntOrStrPtr(val int) *intstr.IntOrString {
	v := intstr.FromInt(val)
	return &v
}
