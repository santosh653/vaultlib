package vaultlib

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/pkg/errors"
)

var (
	renewalLoopRunning bool = false
)

// vaultAuth holds the Vault Auth response from server
type vaultAuth struct {
	ClientToken string   `json:"client_token"`
	Accessor    string   `json:"accessor"`
	Policies    []string `json:"policies"`
	Metadata    struct {
		RoleName string `json:"role_name"`
	} `json:"metadata"`
	LeaseDuration int    `json:"lease_duration"`
	Renewable     bool   `json:"renewable"`
	EntityID      string `json:"entity_id"`
}

// renew the client's token, launched at client creation time as a go routine
func (c *Client) renewToken() {

	// Don't run if loop is already running
	if renewalLoopRunning {
		fmt.Println("Renewal loop already running. Skipping.")
		return
	}

	var vaultData vaultAuth
	jsonToken := make(map[string]string)
	renewalLoopRunning = true

	for {
		// Subtract 5% from the TTL so we renew before it expires
		duration := c.token.TTL - int(math.Floor((float64(c.token.TTL) * 0.05)))

		// If token TTL is 0, then don't try renewing
		if duration <= 0 {
			fmt.Println("Token TTL is 0, so renewal loop would run infinitely. Skipping.")
			renewalLoopRunning = false
			return
		}

		time.Sleep(time.Second * time.Duration(duration))

		url := c.address.String() + "/v1/auth/token/renew-self"

		req, _ := c.newRequest("POST", url)

		// Sending a payload (even empty) is required for vault to respond with the `auth` param
		_ = req.setJSONBody(jsonToken)

		resp, err := req.execute()
		if err != nil {
			c.setStatus("Error renewing token " + err.Error())
		}

		if err == nil {
			jsonErr := json.Unmarshal([]byte(resp.Auth), &vaultData)
			if jsonErr != nil {
				c.setStatus("Error renewing token " + err.Error())
				continue
			}
		}

		if err := c.setTokenInfo(); err != nil {
			c.setStatus("Error renewing token " + err.Error())

			// Try logging in again. If that fails, crash.
			err := c.setTokenFromAppRole()
			if err != nil {
				renewalLoopRunning = false
				panic(err)
			}

			c.setStatus("token ready (new)")

			continue
		}
		c.setStatus("token renewed")
	}
}

// setTokenFromAppRole get the token from Vault and set it in the client
func (c *Client) setTokenFromAppRole() error {
	var vaultData vaultAuth

	mp := "approle"
	if c.appRoleCredentials.MountPoint != "" {
		mp = c.appRoleCredentials.MountPoint
	}

	if c.appRoleCredentials.RoleID == "" || c.appRoleCredentials.SecretID == "" {
		return errors.New("No credentials provided")
	}

	url := fmt.Sprintf("%s/v1/auth/%s/login", c.address.String(), mp)

	req, _ := c.newRequest("POST", url)

	_ = req.setJSONBody(c.appRoleCredentials)

	resp, err := req.execute()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), errInfo())
	}

	jsonErr := json.Unmarshal([]byte(resp.Auth), &vaultData)
	if jsonErr != nil {
		return errors.Wrap(errors.WithStack(err), errInfo())
	}
	c.withLockContext(func() {
		c.token.ID = vaultData.ClientToken
	})

	if err = c.setTokenInfo(); err != nil {
		return errors.Wrap(errors.WithStack(err), errInfo())
	}
	if c.token.Renewable && !renewalLoopRunning {
		go c.renewToken()
	}

	c.setStatus("token ready")

	return nil
}

// vaultSecretKV2 holds the Vault secret (kv v2)
type vaultSecretKV2 struct {
	Data     map[string]interface{} `json:"data"`
	Metadata struct {
		CreatedTime  time.Time `json:"created_time"`
		DeletionTime string    `json:"deletion_time"`
		Destroyed    bool      `json:"destroyed"`
		Version      int       `json:"version"`
	} `json:"metadata"`
}

func (c *Client) setTokenInfo() error {
	url := c.address.String() + "/v1/auth/token/lookup-self"
	var tokenInfo VaultTokenInfo

	req, _ := c.newRequest("GET", url)

	res, err := req.execute()
	if err != nil {
		return err
	}
	if err := json.Unmarshal(res.Data, &tokenInfo); err != nil {
		return err
	}
	c.withLockContext(func() {
		c.token = &tokenInfo
		c.isAuthenticated = true
		c.status = "token ready"
	})
	return nil
}
