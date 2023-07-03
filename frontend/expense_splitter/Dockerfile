ARG USER_ID=1000
ARG GROUP_ID=1000

FROM node:lts-alpine AS builder
ARG USER_ID
ARG GROUP_ID

RUN apk --no-cache add curl git make musl-dev go helm coreutils &&\
    curl -fsSL "https://github.com/pnpm/pnpm/releases/latest/download/pnpm-linuxstatic-x64" -o /bin/pnpm &&\
    chmod +x /bin/pnpm &&\
    (addgroup -S ${GROUP_ID} || echo "Group ${GROUP_ID} already exists.") &&\
    (adduser -S ${USER_ID} -G ${GROUP_ID} -u ${USER_ID} || echo "User ${USER_ID} already exists.") &&\
    mkdir -p /usr/src/app/frontend/expense_splitter &&\
    chown -R ${USER_ID}:${GROUP_ID} /usr/src/app
USER ${USER_ID}
WORKDIR /usr/src/app

COPY --chown=${USER_ID}:${GROUP_ID} Makefile pnpm-lock.yaml pnpm-workspace.yaml /usr/src/app/
COPY --chown=${USER_ID}:${GROUP_ID} frontend/expense_splitter/package.json /usr/src/app/frontend/expense_splitter/
RUN make pnpm-install
COPY --chown=${USER_ID}:${GROUP_ID} buf.gen.yaml.tpl buf.work.yaml /usr/src/app/
COPY --chown=${USER_ID}:${GROUP_ID} proto /usr/src/app/proto
RUN PATH="$PATH:$(eval echo '~/go/bin')" make generate-proto-with-node &&\
    rm -rf gen/doc &&\
    rm -rf gen/lib/go &&\
    rm -rf '~/go'
COPY --chown=${USER_ID}:${GROUP_ID} ./frontend/expense_splitter /usr/src/app/frontend/expense_splitter
WORKDIR /usr/src/app/frontend/expense_splitter
RUN pnpm build

FROM node:lts-alpine
ARG USER_ID
ARG GROUP_ID

RUN (addgroup -S ${GROUP_ID} || echo "Group ${GROUP_ID} already exists.") &&\
    (adduser -S ${USER_ID} -G ${GROUP_ID} -u ${USER_ID} || echo "User ${USER_ID} already exists.") &&\
    mkdir -p /usr/src/app/frontend/expense_splitter/build &&\
    chown -R ${USER_ID}:${GROUP_ID} /usr/src/app
USER ${USER_ID}
WORKDIR /usr/src/app/frontend/expense_splitter
COPY --from=builder --chown=${USER_ID}:${GROUP_ID} /usr/src/app/frontend/expense_splitter/package.json /usr/src/app/frontend/expense_splitter/package.json
COPY --from=builder --chown=${USER_ID}:${GROUP_ID} /usr/src/app/frontend/expense_splitter/node_modules /usr/src/app/frontend/expense_splitter/node_modules
COPY --from=builder --chown=${USER_ID}:${GROUP_ID} /usr/src/app/frontend/expense_splitter/build /usr/src/app/frontend/expense_splitter/build

ENTRYPOINT [ "node", "build/index.js" ]