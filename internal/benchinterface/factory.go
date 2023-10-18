package benchinterface

import (
	config "serverlessbench2/internal/config"
)

type Provider interface {
	CreateFunction(functionName string, memorySize int, sourceCodeFilePath ...string) error
	CreateWorkflow(flowName string, functionInfo ...map[string](map[string]string)) error
	DeleteFunction(functionName string) error
	DeleteWorkflow(flowName string) error
	ListFunction() error
	InvokeFunction(functionName string, params ...string) error
	InvokeWorkflow(functionName string, params ...string) error
}

func CheckProvider() string {
	return config.GetProvider()
}

func providerFactory() Provider {
	p := CheckProvider()
	switch p {
	case "knative":
		return ProviderKnative{}
	case "openwhisk":
		return ProviderOpenwhisk{}
	case "openfaas":
		return ProviderOpenfaas{}
	case "fission":
		return ProviderFission{}
	default:
		return nil
	}
}
