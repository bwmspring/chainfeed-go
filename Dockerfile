FROM golang:1.20-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/chainfeed ./cmd/server

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/chainfeed /bin/chainfeed
ENV SERVER_PORT=8080
EXPOSE 8080
ENTRYPOINT ["/bin/chainfeed"]
