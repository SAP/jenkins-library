FROM golang:1.13 AS build-env
COPY . /build
WORKDIR /build

# execute tests
RUN go test ./... -cover

## ONLY tests so far, building to be added later
# execute build
# RUN go build -o piper
RUN export GIT_COMMIT=$(git rev-parse HEAD) && \
    go build -ldflags "-X github.com/SAP/jenkins-library/cmd.GitCommit=${GIT_COMMIT}" -o piper

# FROM gcr.io/distroless/base:latest
# COPY --from=build-env /build/piper /piper
# ENTRYPOINT ["/piper"]
