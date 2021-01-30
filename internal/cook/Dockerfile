# GERENATE BUILDER
FROM golang:1.15.5-buster as builder

ARG SSH_HOST
ARG PRIVATE_KEY
ARG FINGER_PRINT

WORKDIR /go/src/github.com/danielsussa/floxy/internal/cook/
COPY . /go/src/github.com/danielsussa/floxy/internal/cook/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GODEBUG=madvdontneed=1 go build \
 -ldflags "-X main.Kind=docker -X main.FingerPrint=$FINGER_PRINT -X main.PrivateKey=$PRIVATE_KEY -X main.SshHost=$SSH_HOST" -a -tags netgo  -o /build/application cook.go

RUN ls

# CREATE FINAL DOCKER IMAGE

FROM golang:1.15-alpine as runner

RUN apk --no-cache add \
    ca-certificates \
    curl \
    bash \
    tzdata

WORKDIR /
COPY --from=builder /build/application .
EXPOSE 8080
ENTRYPOINT ["./application"]