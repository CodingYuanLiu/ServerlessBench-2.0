package config

import "fmt"

type configProvider interface {
	writeConfig() error
	checkConfig() error
}

func providerFactory() configProvider {
	p := GetProvider()
	switch p {
	case "knative":
		return knativeProvider{}
	case "openwhisk":
		return openwhiskProvider{}
	case "openfaas":
		return openfaasProvider{}
	case "fission":
		return fissionProvider{}
	default:
		return nil
	}
}

func WriteConfig(BaseConfig *BaseConfig) error {
	err := writeBaseConfig(BaseConfig)
	if err != nil {
		return err
	}
	provider := providerFactory()
	if provider != nil {
		err = provider.writeConfig()
		if err != nil {
			return err
		}
		err = provider.checkConfig()
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("Warning: provider not configured(only support knative, openwhisk, openfaas, fission)\n")
	}
	return nil
}

type knativeProvider struct {
	configProvider
}

type openwhiskProvider struct {
	configProvider
}

type openfaasProvider struct {
	configProvider
}

type fissionProvider struct {
	configProvider
}
