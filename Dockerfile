# syntax=docker/dockerfile:1.7

FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /out/tacticboard ./cmd/server

FROM gcr.io/distroless/base-debian12:nonroot
ENV APP_ENV=production \
    APP_HOST=0.0.0.0 \
    APP_PORT=8080

COPY --from=builder /out/tacticboard /bin/tacticboard
EXPOSE 8080

ENTRYPOINT ["/bin/tacticboard"]
