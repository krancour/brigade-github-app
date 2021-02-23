module github.com/brigadecore/brigade-github-app/v2

go 1.15

replace k8s.io/client-go => k8s.io/client-go v0.18.2

require (
	github.com/brigadecore/brigade/sdk/v2 v2.0.0-alpha.1
	github.com/brigadecore/brigade/v2 v2.0.0-alpha.1
	github.com/gin-gonic/gin v1.5.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-github/v32 v32.0.0
	github.com/gorilla/mux v1.8.0
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.7.0
	github.com/stretchr/testify v1.6.1
	gopkg.in/gin-gonic/gin.v1 v1.1.5-0.20170702092826-d459835d2b07
	gopkg.in/go-playground/validator.v9 v9.31.0 // indirect
)
