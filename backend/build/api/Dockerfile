FROM golang:1.22 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY internal/ ./internal/
COPY cmd/ ./cmd/
RUN go build -o /out/api ./cmd/api

FROM gcr.io/distroless/base-debian12:latest
WORKDIR /app
COPY --from=build /out/api ./api
COPY config.yaml ./
COPY web ./
USER nonroot:nonroot
ENTRYPOINT ["./api", "--config", "./config.yaml"]
