package main

import (
	"log"

	"github.com/brigadecore/brigade-github-app/v2/gateway/internal/lib/restmachinery"
	"github.com/brigadecore/brigade-github-app/v2/gateway/internal/webhooks"
	"github.com/brigadecore/brigade-github-app/v2/gateway/internal/webhooks/rest"
	"github.com/brigadecore/brigade-github-app/v2/internal/signals"
	"github.com/brigadecore/brigade-github-app/v2/internal/version"
	"github.com/brigadecore/brigade/sdk/v2/core"
)

func main() {

	log.Printf(
		"Starting Brigade GitHub App -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	address, token, opts, err := apiClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	authFilterConfig, err := authFilterConfig()
	if err != nil {
		log.Fatal(err)
	}

	serverConfig, err := serverConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(
		restmachinery.NewServer(
			[]restmachinery.Endpoints{
				&rest.WebhookEndpoints{
					AuthFilter: rest.NewAuthFilter(authFilterConfig),
					Service: webhooks.NewService(
						core.NewEventsClient(address, token, &opts),
						webhookServiceConfig(),
					),
				},
			},
			&serverConfig,
		).ListenAndServe(signals.Context()),
	)
}
