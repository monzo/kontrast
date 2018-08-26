package k8s

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

var metadataAccessor = meta.NewAccessor()

type Resource struct {
	Name      string
	Namespace string
	Object    runtime.Object
	helper    *ResourceHelper
}

func (r *Resource) RESTMapping(m meta.RESTMapper) (*meta.RESTMapping, error) {
	gvk := r.Object.GetObjectKind().GroupVersionKind()
	return m.RESTMapping(gvk.GroupKind(), gvk.GroupVersion().Version)
}

func (r *Resource) Get() (runtime.Object, error) {
	return r.helper.Get(r)
}

func (r *Resource) Create() error {
	return r.helper.Create(r)
}

func (r *Resource) Delete() error {
	return r.helper.Delete(r)
}
