package k8s

import (
	"fmt"
	"io/ioutil"
	"log"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type ResourceHelper struct {
	Config *rest.Config
	meta.RESTMapper
	DefaultNamespace string
	Scheme           *runtime.Scheme
}

func NewResourceHelperWithDefaults(config *rest.Config) (*ResourceHelper, error) {
	return NewResourceHelper(config, "default")
}

func NewResourceHelper(config *rest.Config, defaultNamespace string) (*ResourceHelper, error) {
	client, err := rest.UnversionedRESTClientFor(config)
	if err != nil {
		return &ResourceHelper{}, fmt.Errorf("create REST client: %s", err.Error())
	}

	discoveryClient := discovery.NewDiscoveryClient(client)
	apiGroupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return &ResourceHelper{}, fmt.Errorf("discover APIGroupResources: %s", err.Error())
	}

	mapper := restmapper.NewDiscoveryRESTMapper(apiGroupResources)

	return &ResourceHelper{
		Config:           config,
		RESTMapper:       mapper,
		DefaultNamespace: defaultNamespace,
		Scheme:           scheme.Scheme,
	}, nil
}

func (rh *ResourceHelper) NewResourceFromFilename(filename string) (*Resource, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return &Resource{}, fmt.Errorf("read file %s: %s", filename, err.Error())
	}
	return rh.NewResourceFromBytes(bytes)
}

func (rh *ResourceHelper) NewResourceFromBytes(bytes []byte) (*Resource, error) {
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(bytes, nil, nil)
	if err != nil {
		return &Resource{}, fmt.Errorf("parse resource from bytes: %s", err.Error())
	}
	return rh.NewResource(obj)
}

func (rh *ResourceHelper) NewResource(obj runtime.Object) (*Resource, error) {
	name, _ := metadataAccessor.Name(obj)
	namespace, _ := metadataAccessor.Namespace(obj)

	if namespace == "" {
		namespace = rh.DefaultNamespace
	}

	return &Resource{
		Name:      name,
		Namespace: namespace,
		Object:    obj,
		helper:    rh,
	}, nil
}

func (rh *ResourceHelper) mapping(gvk schema.GroupVersionKind) (*meta.RESTMapping, error) {
	return rh.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.GroupVersion().Version)
}

func (rh *ResourceHelper) clientFor(gvk schema.GroupVersionKind) (rest.Interface, error) {
	config := *rh.Config
	gv := gvk.GroupVersion()
	config.GroupVersion = &gv
	return rest.RESTClientFor(&config)
}

func (rh *ResourceHelper) Create(r *Resource) error {
	gvk := r.Object.GetObjectKind().GroupVersionKind()
	mappedResource, err := rh.mapping(gvk)
	if err != nil {
		return fmt.Errorf("getting RESTMapping: %s", err.Error())
	}
	client, err := rh.clientFor(gvk)
	if err != nil {
		return fmt.Errorf("creating REST client: %s", err.Error())
	}
	req := client.Post().
		Namespace(r.Namespace).
		Resource(mappedResource.Resource.Resource).
		Body(r.Object)

	res := req.Do()

	if res.Error() != nil {
		log.Printf("%#v", res.Error())
		return fmt.Errorf("making REST request: %s", res.Error())
	}
	return nil
}

func (rh *ResourceHelper) Get(r *Resource) (runtime.Object, error) {
	gvk := r.Object.GetObjectKind().GroupVersionKind()
	mappedResource, err := rh.mapping(gvk)
	if err != nil {
		return &v1.List{}, fmt.Errorf("getting RESTMapping: %s", err.Error())
	}
	client, err := rh.clientFor(gvk)
	if err != nil {
		return &v1.List{}, fmt.Errorf("creating REST client: %s", err.Error())
	}
	req := client.Get().
		Namespace(r.Namespace).
		Resource(mappedResource.Resource.Resource).
		Param("export", "true").
		Name(r.Name)

	res := req.Do()

	if res.Error() != nil {
		return &v1.List{}, res.Error()
	}
	return res.Get()
}

func (rh *ResourceHelper) Delete(r *Resource) error {
	gvk := r.Object.GetObjectKind().GroupVersionKind()
	mappedResource, err := rh.mapping(gvk)
	if err != nil {
		return fmt.Errorf("getting RESTMapping: %s", err.Error())
	}
	client, err := rh.clientFor(gvk)
	if err != nil {
		return fmt.Errorf("creating REST client: %s", err.Error())
	}
	req := client.Delete().
		Namespace(r.Namespace).
		Resource(mappedResource.Resource.Resource).
		Name(r.Name)

	res := req.Do()

	if res.Error() != nil {
		log.Printf("%#v", res.Error())
		return fmt.Errorf("making REST request: %s", res.Error())
	}
	return nil
}

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	st, ok := err.(*errors.StatusError)
	if ok {
		return st.Status().Reason == "NotFound"
	}
	return false
}