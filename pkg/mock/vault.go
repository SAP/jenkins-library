// +build !release

package mock

import (
	"strconv"
	"strings"

	"github.com/hashicorp/vault/api"
)

// VaultClientMock implements the functions from vault.logicalClient with an in-memory secret store
type VaultClientMock struct {
	store     map[string]*api.Secret
	kvVersion int
}

func (v *VaultClientMock) init() {
	if v.store == nil {
		v.store = map[string]*api.Secret{}
	}
	if v.kvVersion == 0 {
		v.kvVersion = 2
	}
}

// SetKvEngineVersion allows to toggle the version of the virtual KV Engine
func (v *VaultClientMock) SetKvEngineVersion(version int) {
	if version != 1 && version != 2 {
		v.kvVersion = 2
	}
	v.kvVersion = version
}

// AddSecret establishes a virtual secret which then can be retrieved by the client
func (v *VaultClientMock) AddSecret(path string, secret *api.Secret) {
	v.init()
	if secret == nil {
		return
	}
	v.store[path] = secret
}

func (v *VaultClientMock) Read(path string) (*api.Secret, error) {
	v.init()
	if strings.HasPrefix(path, "sys/internal/ui/mounts/") {
		pathComponents := strings.Split(strings.TrimPrefix(path, "sys/internal/ui/mounts/"), "/")
		mountpath := "/"
		if len(pathComponents) > 1 {
			mountpath = pathComponents[0]
		}
		switch v.kvVersion {
		case 1:
			// in older versions of vault the options field was not present
			return &api.Secret{
				Data: map[string]interface{}{
					"path": mountpath,
				},
			}, nil
		default:
			return &api.Secret{
				Data: map[string]interface{}{
					"path": mountpath,
					"options": map[string]interface{}{
						"version": strconv.Itoa(v.kvVersion),
					},
				},
			}, nil
		}
	}
	if secret, ok := v.store[path]; ok {
		return secret, nil
	}
	return nil, nil
}
