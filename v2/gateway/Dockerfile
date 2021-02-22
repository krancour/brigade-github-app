FROM brigadecore/go-tools:v0.1.0
ARG VERSION
ARG COMMIT
ENV CGO_ENABLED=0
WORKDIR /src/v2
COPY v2/gateway/ gateway/
COPY v2/internal/ internal/
COPY v2/go.mod go.mod
COPY v2/go.sum go.sum

RUN go build \
  -o ../bin/gateway \
  -ldflags "-w -X github.com/brigadecore/brigade-github-app/v2/internal/version.version=$VERSION -X github.com/brigadecore/brigade-github-app/v2/internal/version.commit=$COMMIT" \
  ./gateway

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /src/bin/ /brigade-github-app/bin/
ENTRYPOINT ["/brigade-github-app/bin/gateway"]
