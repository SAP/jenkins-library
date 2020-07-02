package mock

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

//VaultClientMock implements the functions from vault.logicalClient with an in-memory secret store
type VaultClientMock struct {
	store map[string]*api.Secret
}

func (v *VaultClientMock) init() {
	if v.store == nil {
		v.store = map[string]*api.Secret{}
	}
}

//AddSecret establishes a virtual secret which then can be retrieved by the client
func (v *VaultClientMock) AddSecret(path string, secret *api.Secret) {
	v.init()
	if secret == nil {
		return
	}
	v.store[path] = secret
}

func (v *VaultClientMock) Read(path string) (*api.Secret, error) {
	if secret, ok := v.store[path]; ok {
		return secret, nil
	}
	return nil, fmt.Errorf("No mocked secret for path: %s", path)
}
