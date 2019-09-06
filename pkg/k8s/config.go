package k8s

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// LoadConfig initialises the config required to communicate with the K8s API.
// It will first attempt an in-cluster config, then load from the given
// kubeconfig filename.
func LoadConfig(kubeconfig string) (*rest.Config, error) {
	var config *rest.Config
	log.Info("Trying to load in-cluster config...")
	config, err := rest.InClusterConfig()

	if err != nil {
		log.Warnf("In cluster config not successful, trying out of cluster config (error: %v)", err)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return config, err
		}
		log.Info("Loaded out of cluster config")
	} else {
		log.Info("Loaded in-cluster config")
	}

	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	return config, nil
}
