package btp

import (
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/btputils"
)

func CheckServiceInstanceCreated(btp *BTPUtils, options GetServiceInstanceOptions) bool {
	serviceInstanceJSON, err := btp.GetServiceInstance(options)

	if err != nil {
		fmt.Println("Service Instance not found...")
		return false
	}

	data := btputils.ServiceInstanceData{}

	err = json.Unmarshal([]byte(serviceInstanceJSON), &data)

	if err != nil {
		return false
	}

	err = btp.Logout()
	if err != nil {
		return false
	}

	return data.Ready
}

func CheckServiceInstanceDeleted(btp *BTPUtils, options GetServiceInstanceOptions) bool {
	_, err := btp.GetServiceInstance(options)

	if err == nil {
		fmt.Println("Instance still exists...")
		return false
	}

	return true
}

func CheckServiceBindingCreated(btp *BTPUtils, options GetServiceBindingOptions) bool {
	serviceBindingJSON, err := btp.GetServiceBinding(options)

	if err != nil {
		fmt.Println("Service Binding not found...")
		return false
	}

	data := btputils.ServiceBindingData{}

	err = json.Unmarshal([]byte(serviceBindingJSON), &data)

	if err != nil {
		return false
	}

	err = btp.Logout()
	if err != nil {
		return false
	}

	return data.Ready
}

func CheckServiceBindingDeleted(btp *BTPUtils, options GetServiceBindingOptions) bool {
	_, err := btp.GetServiceBinding(options)

	if err == nil {
		fmt.Println("Binding still exists")
		return false
	}

	return true
}
