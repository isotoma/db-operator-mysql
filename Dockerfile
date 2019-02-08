FROM golang AS builder

# This is unpleasant, but means we can depend on private github
# repos. Once the repos are public, this can all go away. (Note that
# the private-key is copied into the builder container, but isn't
# copied on to the final container).
ARG SSH_PRIVATE_KEY
RUN mkdir /root/.ssh/
RUN echo "${SSH_PRIVATE_KEY}" > /root/.ssh/id_rsa
RUN chmod 400 /root/.ssh/id_rsa
RUN touch /root/.ssh/known_hosts
RUN ssh-keyscan github.com >> /root/.ssh/known_hosts
# (end ssh key configuration)

ENV GOPATH /app
RUN mkdir -p /app/src/db-operator-mysql
RUN mkdir /build
WORKDIR /app/src/db-operator-mysql

RUN curl -fL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64
RUN chmod +x /usr/local/bin/dep

# First, add the dependency files and install dependencies
ADD Gopkg.lock Gopkg.toml ./
RUN dep ensure --vendor-only

# Then add the source and build it. By keeping these steps separate,
# we let docker do more caching and so builds are faster if the deps
# haven't changed.
ADD . .

ENV CGO_ENABLED=0
ENV GOBIN="/app/src/db-operator-mysql/bin"

RUN go install db-operator-mysql/cmd/... && strip bin/* && cp bin/driver /build/db-operator-mysql

######

FROM alpine
COPY --from=builder /build/db-operator-mysql /
CMD /db-operator-mysql
