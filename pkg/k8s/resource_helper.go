package k8s

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

// ResourceHelper manages getting, updating, creating/deleting K8s objects with
// a remote API server
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

	// heavily borrowed from kubectl code; discovers available API groups and
	// creates a mapping between resources and REST mappings (e.g. apps/v1beta2
	// Deployment => /apis/apps/v1beta2/namespaces/<namespace>/deployments)
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

// NewResourcesFromFilename creates Resource wrappers for each manifest found
// in the filename passed in.
// TODO: this struct shouldn't be responsible for files or YAML parsing, we
// should just pass in an object's worth of bytes
func (rh *ResourceHelper) NewResourcesFromFilename(filename string) ([]*Resource, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []*Resource{}, fmt.Errorf("open file %s: %s", filename, err.Error())
	}
	resources := []*Resource{}

	reader := bufio.NewReader(f)
	// use K8s YAML reader to split up documents (1 doc should == 1 object)
	decoder := yaml.NewYAMLReader(reader)

	for {
		bytes, err := decoder.Read()

		if len(bytes) == 0 {
			// no more documents
			return resources, nil
		}

		if err != nil {
			return []*Resource{}, fmt.Errorf("decode doc from %s: %s", filename, err.Error())
		}

		res, err := rh.NewResourceFromBytes(bytes)
		if err != nil {
			return []*Resource{}, fmt.Errorf("deserialise resource %s: %s", filename, err.Error())
		}

		if res != nil {
			resources = append(resources, res)
		}
	}
}

// NewResourceFromBytes creates a new Resource wrapper from the passed in
// bytes. Nothing is done with the remote K8s API server until Resource
// methods are called.
func (rh *ResourceHelper) NewResourceFromBytes(bytes []byte) (*Resource, error) {

	// K8s deserialiser does all the hard work for us here - figures out
	// format, API group, kind, version
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(bytes, nil, nil)

	// A missing `Kind` most probably meant an empty document
	if runtime.IsMissingKind(err) {
		return nil, nil
	}

	// Base64 errors are most often caused by having dummy passwords in Secret files.
	// A `fmt.Errorf` in the call stack removed the type, so this is all there is left.
	// Check for `base64.CorruptInputError` if possible.
	if strings.Contains(err.Error(), "illegal base64 data at input byte") {
		return nil, nil
	}

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

// mapping gets the RESTMapping for this particular object so that the REST
// client can build a URL for this resource - e.g.  apps/v1beta2 Deployment
// => /apis/apps/v1beta2/namespaces/<namespace>/deployments
func (rh *ResourceHelper) mapping(gvk schema.GroupVersionKind) (*meta.RESTMapping, error) {
	return rh.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.GroupVersion().Version)
}

func (rh *ResourceHelper) clientFor(gvk schema.GroupVersionKind) (rest.Interface, error) {

	config := *rh.Config
	gv := gvk.GroupVersion()
	config.GroupVersion = &gv

	// This is crufty :( K8s serves it's "legacy"/"core" API (anything under
	// the v1 prefix) at /api, and all the regular API groups at /apis
	// It makes me feel better that there is a similar piece of code somewhere
	// in the kubectl codebase
	if gv.String() == "v1" {
		config.APIPath = "/api"
	} else {
		config.APIPath = "/apis"
	}

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

func (rh *ResourceHelper) buildGETRequestFor(r *Resource, export bool) (*rest.Request, error) {
	gvk := r.Object.GetObjectKind().GroupVersionKind()

	mappedResource, err := rh.mapping(gvk)
	if err != nil {
		return &rest.Request{}, fmt.Errorf("getting RESTMapping: %s", err.Error())
	}

	client, err := rh.clientFor(gvk)
	if err != nil {
		return &rest.Request{}, fmt.Errorf("creating REST client: %s", err.Error())
	}

	req := client.Get().
		Resource(mappedResource.Resource.Resource).
		Name(r.Name)

	if export {
		req.Param("export", "true")
	}

	if mappedResource.Scope.Name() == "namespace" {
		req.Namespace(r.Namespace)
	}

	return req, nil
}

func (rh *ResourceHelper) Get(r *Resource) (runtime.Object, error) {
	req, err := rh.buildGETRequestFor(r, false)
	if err != nil {
		return &v1.List{}, err
	}
	res := req.Do()

	if res.Error() != nil {
		if strings.HasPrefix(res.Error().Error(), "export of") {
			log.Println("retrying with export disabled")
			req, err := rh.buildGETRequestFor(r, false)
			if err != nil {
				return &v1.List{}, err
			}
			res := req.Do()
			if res.Error() != nil {
				log.Printf("do error:\n%#v\nURL:%s", res.Error(), req.URL().String())
				return &v1.List{}, res.Error()
			}
			obj, err := res.Get()
			if err != nil {
				log.Printf("get error: %#v", res.Error())
			}
			return obj, err
		} else {
			log.Printf("do error:\n%#v\nURL:%s", res.Error(), req.URL().String())
			return &v1.List{}, res.Error()
		}
	}
	obj, err := res.Get()
	if err != nil {
		log.Printf("get error: %#v", res.Error())
	}
	return obj, err
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
	st, ok := err.(*errors.StatusError)
	if ok {
		return st.Status().Reason == "NotFound"
	}
	return false
}
