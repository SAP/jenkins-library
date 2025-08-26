package vault

import (
	"context"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	vaultAPI "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client handles communication with Vault
type Client struct {
	vaultApiClient *vaultAPI.Client
	logical        logicalClient
	cfg            *ClientConfig
}

type ClientConfig struct {
	*vaultAPI.Config
	Namespace         string
	AppRoleMountPoint string
	RoleID            string
	SecretID          string
}

// logicalClient interface for mocking
type logicalClient interface {
	Read(string) (*vaultAPI.Secret, error)
	Write(string, map[string]interface{}) (*vaultAPI.Secret, error)
}

func newClient(cfg *ClientConfig) (*Client, error) {
	if cfg == nil {
		cfg = &ClientConfig{Config: vaultAPI.DefaultConfig()}
	}

	var err error
	c := &Client{cfg: cfg}
	c.vaultApiClient, err = vaultAPI.NewClient(cfg.Config)
	if err != nil {
		return nil, err
	}
	c.logical = c.vaultApiClient.Logical()
	if cfg.Namespace != "" {
		c.vaultApiClient.SetNamespace(cfg.Namespace)
	}

	return c, nil
}

func NewClient(cfg *ClientConfig) (*Client, error) {
	c, err := newClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "vault client initialization failed")
	}
	applyApiClientRetryConfiguration(c.vaultApiClient)

	initialLoginDone := make(chan struct{})
	go c.startTokenLifecycleManager(initialLoginDone) // this goroutine ends with main goroutine
	// wait for initial login or a failure
	<-initialLoginDone

	// In case of a failure, the function returns an unauthorized client, which will cause subsequent requests to fail.
	return c, nil
}

func NewClientWithToken(cfg *ClientConfig, token string) (*Client, error) {
	c, err := newClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "vault client initialization failed")
	}

	c.vaultApiClient.SetToken(token)
	return c, nil
}

func (c *Client) startTokenLifecycleManager(initialLoginDone chan struct{}) {
	defer func() {
		// make sure to close channel to avoid blocking of the caller
		log.Entry().Debugf("exiting Vault token lifecycle manager")
		initialLoginDone <- struct{}{}
		close(initialLoginDone)
	}()

	initialLoginSucceed := false
	retryAttemptDuration := c.vaultApiClient.MinRetryWait()
	for i := 0; i <= c.vaultApiClient.MaxRetries(); i++ {
		if i != 0 {
			log.Entry().WithField("attempt", i).WithField("maxRetries", c.vaultApiClient.MaxRetries()).WithField("retryDelay", retryAttemptDuration.Seconds()).Info("Retrying Vault login")
			time.Sleep(retryAttemptDuration)
		}

		vaultLoginResp, err := c.login()
		if err != nil {
			if i == 0 {
				log.Entry().WithError(err).Warn("Vault authentication failed")
			} else {
				log.Entry().WithError(err).WithField("attempt", i).Warn("Vault authentication retry failed")
			}
			continue
		}
		if !initialLoginSucceed {
			initialLoginDone <- struct{}{}
			initialLoginSucceed = true
		}

		if !vaultLoginResp.Auth.Renewable {
			log.Entry().Debugf("Vault token is not configured to be renewable.")
			return
		}

		tokenErr := c.manageTokenLifecycle(vaultLoginResp)
		if tokenErr != nil {
			log.Entry().Warnf("unable to start managing token lifecycle: %v", err)
			continue
		}
	}
}

// Starts token lifecycle management. Returns only fatal errors as errors,
// otherwise returns nil, so we can attempt login again.
func (c *Client) manageTokenLifecycle(authResp *vaultAPI.Secret) error {
	watcher, err := c.vaultApiClient.NewLifetimeWatcher(&vaultAPI.LifetimeWatcherInput{Secret: authResp})
	if err != nil {
		return fmt.Errorf("unable to initialize new lifetime watcher for renewing auth token: %w", err)
	}

	go watcher.Start()
	defer watcher.Stop()

	for {
		select {
		// `DoneCh` will return if renewal fails, or if the remaining lease
		// duration is under a built-in threshold and either renewing is not
		// extending it or renewing is disabled. In any case, the caller
		// needs to attempt to log in again.
		case err := <-watcher.DoneCh():
			if err != nil {
				log.Entry().Printf("Failed to renew Vault token: %v. Re-attempting login.", err)
				return nil
			}
			// This occurs once the token has reached max TTL.
			log.Entry().Printf("Token can no longer be renewed. Re-attempting login.")
			return nil

		// Successfully completed renewal
		case <-watcher.RenewCh():
			log.Entry().Printf("Vault token successfully renewed")
		}
	}
}

func (c *Client) login() (*vaultAPI.Secret, error) {
	appRoleAuth, err := approle.NewAppRoleAuth(c.cfg.RoleID, &approle.SecretID{FromString: c.cfg.SecretID})
	if err != nil {
		return nil, fmt.Errorf("unable to initialize appRole auth method: %w", err)
	}

	authInfo, err := c.vaultApiClient.Auth().Login(context.Background(), appRoleAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to appRole auth method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	return authInfo, nil
}

func applyApiClientRetryConfiguration(vaultApiClient *vaultAPI.Client) {
	vaultApiClient.SetMinRetryWait(time.Second * 5)
	vaultApiClient.SetMaxRetryWait(time.Second * 90)
	vaultApiClient.SetMaxRetries(3)
	vaultApiClient.SetCheckRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if resp != nil {
			log.Entry().Debugln("Vault response: ", resp.Status, resp.StatusCode, err)
		} else {
			log.Entry().Debugln("Vault response: ", err)
		}

		isEOF := false
		if err != nil && strings.Contains(err.Error(), "EOF") {
			log.Entry().Infoln("isEOF is true")
			isEOF = true
		}

		if err == io.EOF {
			log.Entry().Infoln("err = io.EOF is true")
		}

		retry, err := vaultAPI.DefaultRetryPolicy(ctx, resp, err)

		if err != nil || err == io.EOF || isEOF || retry {
			log.Entry().Infoln("Retrying vault request...")
			return true, nil
		}
		return false, nil
	})
}
