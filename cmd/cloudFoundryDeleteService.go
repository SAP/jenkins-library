package cmd

import (
	"fmt"
	"os/exec"
	"github.com/pkg/errors"
)

func cloudFoundryDeleteService(CloudFoundryDeleteServiceOptions cloudFoundryDeleteServiceOptions) error {
	//CF Login
	if(CloudFoundryDeleteServiceOptions.API == "" || CloudFoundryDeleteServiceOptions.Organisation == "" || CloudFoundryDeleteServiceOptions.Space == "" || CloudFoundryDeleteServiceOptions.Username == "" || CloudFoundryDeleteServiceOptions.Password == "") {
		var err = errors.New("Parameters missing. Please provide EITHER the Cloud Foundry ApiEndpoint, Organization, Space, Username or Password!")
		return err
	}
	cflogin, _ := exec.Command("sh", "-c", "cf login -a "+CloudFoundryDeleteServiceOptions.API+" -o "+CloudFoundryDeleteServiceOptions.Organisation+" -s "+CloudFoundryDeleteServiceOptions.Space+" -u "+CloudFoundryDeleteServiceOptions.Username+" -p "+CloudFoundryDeleteServiceOptions.Password).Output()
	fmt.Printf("%s\n\n", cflogin)

	//Deletion of CF Service
	if(CloudFoundryDeleteServiceOptions.ServiceName == "" ) {
		var err = errors.New("Parameter missing. Please provide the Name of the Service Instance you want to delete!")
		return err
	}
	cfdeleteService, _ := exec.Command("sh", "-c", "cf delete-service "+CloudFoundryDeleteServiceOptions.ServiceName+" -f").Output()
	fmt.Printf("%s\n\n", cfdeleteService)

	//CF Logout
	exec.Command("sh", "-c", "cf logout")
	return nil
}
