FROM golang:1

ENV PROJECT=upp-next-video-content-collection-mapper

ENV ORG_PATH="github.com/Financial-Times"
ENV SRC_FOLDER="${GOPATH}/src/${ORG_PATH}/${PROJECT}"
ENV BUILDINFO_PACKAGE="${ORG_PATH}/${PROJECT}/vendor/${ORG_PATH}/service-status-go/buildinfo."

COPY . ${SRC_FOLDER}
WORKDIR ${SRC_FOLDER}

# Set up our extra bits in the image
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# Install dependancies and build app
RUN $GOPATH/bin/dep ensure -vendor-only

RUN VERSION="version=$(git describe --tag --always 2> /dev/null)" \
    && DATETIME="dateTime=$(date -u +%Y%m%d%H%M%S)" \
    && REPOSITORY="repository=$(git config --get remote.origin.url)" \
    && REVISION="revision=$(git rev-parse HEAD)" \
    && BUILDER="builder=$(go version)" \
    && LDFLAGS="-s -w -X '"${BUILDINFO_PACKAGE}$VERSION"' -X '"${BUILDINFO_PACKAGE}$DATETIME"' -X '"${BUILDINFO_PACKAGE}$REPOSITORY"' -X '"${BUILDINFO_PACKAGE}$REVISION"' -X '"${BUILDINFO_PACKAGE}$BUILDER"'" \
    && CGO_ENABLED=0 go build -a -o /artifacts/${PROJECT} -ldflags="${LDFLAGS}" \
    && echo "Build flags: ${LDFLAGS}"

# Multi-stage build - copy certs and the binary into the image
FROM scratch
WORKDIR /
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /artifacts/* /

CMD [ "/upp-next-video-content-collection-mapper" ]