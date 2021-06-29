# Creates a Docker image for running the service.
ARG name=citysearch
ARG bin=/bin/${name}
ARG src=/go/src/github.com/oskanberg/${name}/

####
# Creates a Docker image with the sources.
FROM golang:1.16-buster AS src

# Copy the sources.
ARG src
COPY . ${src}
WORKDIR ${src}

####
# Creates a Docker image with a static binary of the service.
FROM src as build
ARG bin
ARG name
ARG src

RUN CGO_ENABLED=0 go build \
    -o ${bin} \
    -ldflags '-extldflags "-static"' \
    ./cmd/${name}

####
# Creates a new single layer image with binary to run the service.
FROM scratch as run
ARG bin
ARG src

# Copy some required files for the Go stdlib to work: the
# ca-certificates for SSL and the timezone database. Reuses the
# ones in the build image.
COPY --from=src \
    /usr/local/go/lib/time/zoneinfo.zip \
    /usr/local/go/lib/time/zoneinfo.zip
COPY --from=src \
    /etc/ssl/certs/ca-certificates.crt \
    /etc/ssl/certs/ca-certificates.crt

# Copy the database
COPY --from=build ${src}/cities15000.csv /var/cities/
# Copy the binary
COPY --from=build ${bin} /bin/service

ENTRYPOINT ["/bin/service", "--cities", "/var/cities/cities15000.csv"]
