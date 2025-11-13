FROM golang:1.22 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY internal/ ./internal/
COPY cmd/ ./cmd/
RUN go build -o /out/worker ./cmd/worker

FROM gcr.io/distroless/base-debian12:latest
WORKDIR /app
COPY --from=build /out/worker ./worker
COPY config.yaml ./
USER nonroot:nonroot
ENTRYPOINT ["./worker", "--config", "./config.yaml"]
