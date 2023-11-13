FROM golang:1.21-alpine AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN apk add --no-cache gcc musl-dev && CGO_ENABLED=1 go build -ldflags '-w -s' -o /registryswarm

FROM alpine:latest
WORKDIR /app
COPY --from=builder /registryswarm /app/
ENTRYPOINT ["/app/registryswarm"]
