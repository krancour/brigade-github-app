package main

import (
	serverRM "github.com/brigadecore/brigade-github-app/v2/gateway/internal/lib/restmachinery"
	"github.com/brigadecore/brigade-github-app/v2/gateway/internal/webhooks"
	"github.com/brigadecore/brigade-github-app/v2/gateway/internal/webhooks/rest"
	"github.com/brigadecore/brigade-github-app/v2/internal/os"
	clientRM "github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

func apiClientConfig() (string, string, clientRM.APIClientOptions, error) {
	opts := clientRM.APIClientOptions{}
	address, err := os.GetRequiredEnvVar("API_ADDRESS")
	if err != nil {
		return address, "", opts, err
	}
	token, err := os.GetRequiredEnvVar("API_TOKEN")
	if err != nil {
		return address, token, opts, err
	}
	opts.AllowInsecureConnections, err =
		os.GetBoolFromEnvVar("API_IGNORE_CERT_WARNINGS", false)
	return address, token, opts, err
}

func webhookServiceConfig() webhooks.ServiceConfig {
	return webhooks.ServiceConfig{
		AllowedAuthors: os.GetStringSliceFromEnvVar(
			"ALLOWED_AUTHORS",
			[]string{"COLLABORATOR", "OWNER", "MEMBER"},
		),
		EmittedEvents: os.GetStringSliceFromEnvVar(
			"EMITTED_EVENTS",
			[]string{"*"},
		),
	}
}

func authFilterConfig() (rest.AuthFilterConfig, error) {
	config := rest.AuthFilterConfig{}
	var err error
	config.AppID, err = os.GetRequiredEnvVar("GITHUB_APP_ID")
	if err != nil {
		return config, err
	}
	config.SharedSecret, err = os.GetRequiredEnvVar("GITHUB_APP_SHARED_SECRET")
	return config, err
}

func serverConfig() (serverRM.ServerConfig, error) {
	config := serverRM.ServerConfig{}
	var err error
	config.Port, err = os.GetIntFromEnvVar("GATEWAY_SERVER_PORT", 8080)
	if err != nil {
		return config, err
	}
	config.TLSEnabled, err = os.GetBoolFromEnvVar("TLS_ENABLED", false)
	if err != nil {
		return config, err
	}
	if config.TLSEnabled {
		config.TLSCertPath, err = os.GetRequiredEnvVar("TLS_CERT_PATH")
		if err != nil {
			return config, err
		}
		config.TLSKeyPath, err = os.GetRequiredEnvVar("TLS_KEY_PATH")
		if err != nil {
			return config, err
		}
	}
	return config, nil
}
