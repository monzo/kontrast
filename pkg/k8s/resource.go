package k8s

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

var metadataAccessor = meta.NewAccessor()

// Resource wraps a K8s object and provides helpers to manage remote
// operations. Typically this is created by a ResourceHelper
type Resource struct {
	Name      string
	Namespace string
	Object    runtime.Object
	helper    *ResourceHelper
}

// Gets the latest object configuration/status from the API server
func (r *Resource) Get() (runtime.Object, error) {
	return r.helper.Get(r)
}

// Create creates the object on the API server
func (r *Resource) Create() error {
	return r.helper.Create(r)
}

// Delete removes the object from the API server
func (r *Resource) Delete() error {
	return r.helper.Delete(r)
}
