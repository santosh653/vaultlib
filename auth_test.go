package vaultlib

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestVaultClient_setTokenFromAppRole(t *testing.T) {
	rightURL, _ := url.Parse("http://localhost:8200")
	badURL, _ := url.Parse("https://localhost:8200")
	conf := NewConfig()
	anyMountPoint := "anyMountPoint"
	anyCreds := NewConfig().AppRoleCredentials
	anyCreds.MountPoint = anyMountPoint
	htCli := new(http.Client)
	type fields struct {
		Address            *url.URL
		HTTPClient         *http.Client
		AppRoleCredentials *AppRoleCredentials
		//Config     *Config
		Token  string
		Status string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"tokenKO",
			fields{
				rightURL,
				htCli,
				conf.AppRoleCredentials,
				"bad-token",
				""},
			true},
		{"badUrl",
			fields{
				badURL,
				htCli,
				conf.AppRoleCredentials,
				"bad-token",
				""},
			true},
		{"anyMountPoint",
			fields{
				rightURL,
				htCli,
				anyCreds,
				"bad-token",
				""},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				address:            tt.fields.Address,
				httpClient:         tt.fields.HTTPClient,
				appRoleCredentials: tt.fields.AppRoleCredentials,
				//config:     tt.fields.Config,
				token:  &VaultTokenInfo{ID: tt.fields.Token},
				status: tt.fields.Status,
			}
			if err := c.setTokenFromAppRole(); (err != nil) != tt.wantErr {
				t.Errorf("Client.setTokenFromAppRole() error = %v, wantErr %v", c.token.ID, tt.fields.Token)
			}
		})
	}

	// Renewal test
	t.Run("hardRenewal", func(t *testing.T) {
		c := &Client{
			address:            rightURL,
			httpClient:         htCli,
			appRoleCredentials: &AppRoleCredentials{RoleID: longLivedRoleID, SecretID: longLivedSecretID},
			token:              &VaultTokenInfo{ID: "good-token", Renewable: true},
			status:             "",
		}

		// Initial login
		err := c.setTokenFromAppRole()

		assert.Nil(t, err, "Initial login failed")
		assert.Equalf(t, "token ready", c.GetStatus(), "Token init failure")

		// Wait for refresh cycle
		time.Sleep(time.Second * time.Duration(c.token.TTL))

		assert.Equal(t, "token ready (new)", c.GetStatus(), "Token renewal mismatch")
	})
}
