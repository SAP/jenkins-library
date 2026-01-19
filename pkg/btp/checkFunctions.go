package btp

import (
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/log"
)

func IsServiceInstanceCreated(btp *BTPUtils, options GetServiceInstanceOptions) CheckResponse {
	serviceInstanceJSON, err := btp.RunGetServiceInstance(options)

	if err != nil {
		log.Entry().Infof("Service Instance %v not found...", options.InstanceName)
		return CheckResponse{successful: false, done: false}
	}

	data := ServiceInstanceData{}

	err = json.Unmarshal([]byte(serviceInstanceJSON), &data)

	if err != nil {
		log.Entry().Errorf("Parsing of service instance JSON failed: %v", err)
		return CheckResponse{successful: false, done: false}
	}

	if data.Ready {
		log.Entry().Infof("Service Instance %v is ready.", options.InstanceName)
	} else {
		log.Entry().Infof("Service Instance %v is not ready yet.", options.InstanceName)
	}

	return CheckResponse{successful: true, done: data.Ready}
}

func IsServiceInstanceDeleted(btp *BTPUtils, options GetServiceInstanceOptions) CheckResponse {
	_, err := btp.RunGetServiceInstance(options)

	if err == nil {
		log.Entry().Infof("Service Instance %v still exists...", options.InstanceName)
		return CheckResponse{successful: false, done: false}
	}

	log.Entry().Infof("Service Instance %v deleted!", options.InstanceName)
	return CheckResponse{successful: true, done: true}
}

func IsServiceBindingCreated(btp *BTPUtils, options GetServiceBindingOptions) CheckResponse {
	serviceBindingJSON, err := btp.RunGetServiceBinding(options)

	if err != nil {
		log.Entry().Infof("Service Binding %v not found...", options.BindingName)
		return CheckResponse{successful: false, done: false}
	}

	data := ServiceBindingData{}

	err = json.Unmarshal([]byte(serviceBindingJSON), &data)

	if err != nil {
		log.Entry().Errorf("Parsing of service binding JSON failed: %v", err)
		return CheckResponse{successful: true, done: false}
	}

	if data.Ready {
		log.Entry().Infof("Service Binding %v is ready.", options.BindingName)
	} else {
		log.Entry().Infof("Service Binding %v is not ready yet.", options.BindingName)
	}

	return CheckResponse{successful: true, done: data.Ready}
}

func IsServiceBindingDeleted(btp *BTPUtils, options GetServiceBindingOptions) CheckResponse {
	_, err := btp.RunGetServiceBinding(options)

	if err == nil {
		log.Entry().Infof("Service Binding %v still exists", options.BindingName)
		return CheckResponse{successful: false, done: false}
	}

	log.Entry().Infof("Service Binding %v deleted!", options.BindingName)
	return CheckResponse{successful: true, done: true}
}

type CheckResponse struct {
	successful bool
	done       bool
}
