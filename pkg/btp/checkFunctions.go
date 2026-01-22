package btp

import (
	"encoding/json"
)

func IsServiceInstanceCreated(btp *BTPUtils, options GetServiceInstanceOptions) CheckResponse {
	serviceInstanceJSON, err := btp.RunGetServiceInstance(options)

	if err != nil {
		return CheckResponse{successful: false, done: false}
	}

	data := ServiceInstanceData{}

	err = json.Unmarshal([]byte(serviceInstanceJSON), &data)

	if err != nil {
		return CheckResponse{successful: false, done: false}
	}

	return CheckResponse{successful: true, done: data.Ready}
}

func IsServiceInstanceDeleted(btp *BTPUtils, options GetServiceInstanceOptions) CheckResponse {
	_, err := btp.RunGetServiceInstance(options)

	if err == nil {
		return CheckResponse{successful: false, done: false}
	}

	return CheckResponse{successful: true, done: true}
}

func IsServiceBindingCreated(btp *BTPUtils, options GetServiceBindingOptions) CheckResponse {
	serviceBindingJSON, err := btp.RunGetServiceBinding(options)

	if err != nil {
		return CheckResponse{successful: false, done: false}
	}

	data := ServiceBindingData{}

	err = json.Unmarshal([]byte(serviceBindingJSON), &data)

	if err != nil {
		return CheckResponse{successful: true, done: false}
	}

	return CheckResponse{successful: true, done: data.Ready}
}

func IsServiceBindingDeleted(btp *BTPUtils, options GetServiceBindingOptions) CheckResponse {
	_, err := btp.RunGetServiceBinding(options)

	if err == nil {
		return CheckResponse{successful: false, done: false}
	}

	return CheckResponse{successful: true, done: true}
}

type CheckResponse struct {
	successful bool
	done       bool
}
