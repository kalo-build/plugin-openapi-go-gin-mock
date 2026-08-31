package generate

import (
	"fmt"
	"strings"
)

// GenerateDockerfile produces a multi-stage Dockerfile for the generated mock server.
func GenerateDockerfile(cfg Config) string {
	var b strings.Builder
	b.WriteString("FROM golang:1.21-alpine AS builder\n\n")
	b.WriteString("WORKDIR /app\n")
	b.WriteString("COPY go.mod ./\n")
	b.WriteString("RUN go mod download\n\n")
	b.WriteString("COPY . .\n")
	b.WriteString("RUN CGO_ENABLED=0 go build -o /server ./cmd/server\n\n")
	b.WriteString("FROM alpine:3.19\n")
	b.WriteString("RUN apk --no-cache add ca-certificates\n")
	b.WriteString("COPY --from=builder /server /server\n\n")
	b.WriteString(fmt.Sprintf("EXPOSE %s\n", cfg.Port))
	b.WriteString("CMD [\"/server\"]\n")
	return b.String()
}
