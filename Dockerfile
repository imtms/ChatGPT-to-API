FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:alpine AS builder
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
LABEL stage=gobuilder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /app/chatgpt-to-api .

FROM --platform=${TARGETPLATFORM:-linux/amd64} scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
WORKDIR /app
COPY --from=builder /app/chatgpt-to-api /app/chatgpt-to-api
EXPOSE 8080
CMD ["./chatgpt-to-api"]