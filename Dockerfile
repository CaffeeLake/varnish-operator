# Assumes you will not have access to `dep` in the environment.
FROM instrumentisto/dep:0.5-alpine AS builder

# Copy in the go src
WORKDIR /go/src/icm-varnish-k8s-operator
COPY pkg/       pkg/
COPY cmd/       cmd/
COPY Gopkg.toml Gopkg.lock ./

# Populate the vendor folder
RUN dep ensure -v

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager icm-varnish-k8s-operator/cmd/manager

FROM alpine:3.7
LABEL maintainer="thurston sandberg <thurston.sandberg@us.ibm.com>"

RUN apk update &&\
    apk upgrade

COPY --from=builder /go/src/icm-varnish-k8s-operator/manager /manager

USER controller
# RUN chown -R controller
ENTRYPOINT ["/manager"]
