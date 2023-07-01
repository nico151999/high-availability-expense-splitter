# the name of the service (must equal the postfix of the Makefile target for building it)
ARG SERVICE_NAME='my-service'
# the name of the MAKEFILE parameter defining the service's output directory
ARG SVC_OUT_DIR_PARAM='MY_SERVICE_OUT_DIR'
ARG BUILD_PATH="/tmp/${SERVICE_NAME}-service"
ARG BUILD_TARGET="${BUILD_PATH}/${SERVICE_NAME}"
ARG MAKE_TARGET_BUILD="build-${SERVICE_NAME}-service"
ARG GROUP_ID=1000
ARG USER_ID=1000


FROM golang:1.20-alpine AS build_base
ARG BUILD_PATH
ARG BUILD_TARGET
ARG MAKE_TARGET_BUILD
ARG SVC_OUT_DIR_PARAM

RUN apk add --no-cache git make curl coreutils

WORKDIR ${BUILD_PATH}

COPY Makefile ${BUILD_PATH}/
# install tools required to compile; this is not necessary to be done here already but it can improve build times due to caching
RUN make install-buf install-gomplate

COPY go.mod go.sum ${BUILD_PATH}/
RUN go mod download

COPY buf.gen.yaml.tpl buf.work.yaml ${BUILD_PATH}/
COPY proto ${BUILD_PATH}/proto
RUN make generate-proto SKIP_TS_PROTO_INITIALIZATION=true

COPY cmd ${BUILD_PATH}/cmd
COPY internal ${BUILD_PATH}/internal
COPY pkg ${BUILD_PATH}/pkg
RUN --mount=type=cache,target=~/.cache/go-build \
    make ${MAKE_TARGET_BUILD} ${SVC_OUT_DIR_PARAM}=${BUILD_TARGET}


# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
ARG BUILD_TARGET
ARG GROUP_ID
ARG USER_ID

COPY --from=build_base ${BUILD_TARGET} /app/server
USER ${USER_ID}:${GROUP_ID}

EXPOSE 8080

ENTRYPOINT ["/app/server"]