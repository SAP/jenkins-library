package btp

import (
	"encoding/json"
	"fmt"
)

func IsServiceInstanceCreated(btp *BTPUtils, options GetServiceInstanceOptions) bool {
	serviceInstanceJSON, err := btp.GetServiceInstance(options)

	if err != nil {
		fmt.Println("Service Instance not found...")
		return false
	}

	data := ServiceInstanceData{}

	err = json.Unmarshal([]byte(serviceInstanceJSON), &data)

	if err != nil {
		return false
	}

	return data.Ready
}

func IsServiceInstanceDeleted(btp *BTPUtils, options GetServiceInstanceOptions) bool {
	_, err := btp.GetServiceInstance(options)

	if err == nil {
		fmt.Println("Service Instance still exists...")
		return false
	}

	fmt.Println("Service Instance deleted!")
	return true
}

func IsServiceBindingCreated(btp *BTPUtils, options GetServiceBindingOptions) bool {
	serviceBindingJSON, err := btp.GetServiceBinding(options)

	if err != nil {
		fmt.Println("Service Binding not found...")
		return false
	}

	data := ServiceBindingData{}

	err = json.Unmarshal([]byte(serviceBindingJSON), &data)

	if err != nil {
		return false
	}

	return data.Ready
}

func IsServiceBindingDeleted(btp *BTPUtils, options GetServiceBindingOptions) bool {
	_, err := btp.GetServiceBinding(options)

	if err == nil {
		fmt.Println("Service Binding still exists")
		return false
	}

	fmt.Println("Service Binding deleted!")
	return true
}
