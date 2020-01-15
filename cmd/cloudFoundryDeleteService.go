package cmd

import (
	"fmt"
	"os/exec"
)

func cloudFoundryDeleteService(CloudFoundryDeleteServiceOptions cloudFoundryDeleteServiceOptions) error {
	exec.Command("sh", "-c", "cf login -a "+CloudFoundryDeleteServiceOptions.API+" -u "+CloudFoundryDeleteServiceOptions.Username+" -p "+CloudFoundryDeleteServiceOptions.Password+" -o "+CloudFoundryDeleteServiceOptions.Organisation+" -s "+CloudFoundryDeleteServiceOptions.Space).Output()

	//Deletion of CF Service
	cfdeleteService, _ := exec.Command("sh", "-c", "cf delete-service "+CloudFoundryDeleteServiceOptions.ServiceName+" -f").Output()
	fmt.Printf("%s\n\n", cfdeleteService)

    exec.Command("sh", "-c", "cf logout")
    
    
	return nil
}
