package client

import (
	"github.com/rotisserie/eris"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func NewInClusterDynamicClient() (*dynamic.DynamicClient, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, eris.Wrap(err, "failed getting in-cluster config")
	}
	return NewDynamicClient(config)
}

func NewDynamicClient(config *rest.Config) (*dynamic.DynamicClient, error) {
	// creates the dynamic client
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, eris.Wrap(err, "failed creating client set from config")
	}
	return dynClient, nil
}
