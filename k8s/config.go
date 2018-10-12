package k8s

import (
	"log"

	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func LoadConfig(kubeconfigFilename string) (*rest.Config, error) {
	var config *rest.Config
	log.Println("Trying to load in-cluster config...")
	config, err := rest.InClusterConfig()

	if err != nil {
		log.Printf("In cluster config not successful, trying out of cluster config (error: %v", err)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigFilename)
		if err != nil {
			log.Fatalf("Could not get out of cluster config: %v", err)
		}
		log.Println("Loaded out of cluster config")
	}

	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	return config, nil
}
