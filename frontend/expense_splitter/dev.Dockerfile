ARG USER_ID=1000
ARG GROUP_ID=1000

FROM node:lts-alpine
ARG USER_ID
ARG GROUP_ID
ARG BIN_INSTALL_DIR=/usr/src/app/gen/bin

RUN apk --no-cache add curl make coreutils &&\
    (addgroup -S $GROUP_ID || echo "Group $GROUP_ID already exists.") &&\
    (adduser -S $USER_ID -G $GROUP_ID -u $USER_ID || echo "User $USER_ID already exists.") &&\
    mkdir -p /usr/src/app/frontend/expense_splitter &&\
    chown -R $USER_ID:$GROUP_ID /usr/src/app
USER $USER_ID
WORKDIR /usr/src/app

COPY --chown=$USER_ID:$GROUP_ID Makefile /usr/src/app/
RUN make install-pnpm install-gomplate BIN_INSTALL_DIR=$BIN_INSTALL_DIR
COPY --chown=$USER_ID:$GROUP_ID pnpm-lock.yaml pnpm-workspace.yaml /usr/src/app/
COPY --chown=$USER_ID:$GROUP_ID frontend/expense_splitter/package.json /usr/src/app/frontend/expense_splitter/
RUN make pnpm-install
COPY --chown=$USER_ID:$GROUP_ID buf.gen.yaml.tpl buf.gen.tag.yaml.tpl buf.work.yaml /usr/src/app/
COPY --chown=$USER_ID:$GROUP_ID proto /usr/src/app/proto
RUN make generate-proto-with-node &&\
    rm -rf gen/doc &&\
    rm -rf gen/lib/go
COPY --chown=$USER_ID:$GROUP_ID ./frontend/expense_splitter /usr/src/app/frontend/expense_splitter
WORKDIR /usr/src/app/frontend/expense_splitter

ENV PATH="$PATH:$BIN_INSTALL_DIR"
ENTRYPOINT [ "pnpm", "dev", "--host" ]