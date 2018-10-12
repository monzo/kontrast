package k8s

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	// This loads a package which in turn loads every single API group included
	// with the linked K8s version. By loading these, the Scheme.Defaults are
	// populated for all objects. Unfortunately this has the side effect of
	// loading most of the K8s code base, but I can't find a better way of
	// doing this 🤷‍♂️
	_ "k8s.io/kubernetes/pkg/master"
)

// GetWithDefaults returns a copy of the given object, with defaults applied
// (as the K8s API server would do)
func GetWithDefaults(obj runtime.Object) runtime.Object {
	copy := obj.DeepCopyObject()
	legacyscheme.Scheme.Default(copy)
	return copy
}
