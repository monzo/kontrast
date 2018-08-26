package k8s

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	_ "k8s.io/kubernetes/pkg/master"
)

func GetWithDefaults(obj runtime.Object) runtime.Object {
	copy := obj.DeepCopyObject()
	legacyscheme.Scheme.Default(copy)
	return copy
}
