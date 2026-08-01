FROM golang:1.25.2 AS builder

ARG GOOS=linux
ARG CGO_ENABLED=0

WORKDIR /app
COPY . ./

RUN apt-get update \
	&& apt-get install -y ca-certificates git \
	&& rm -rf /var/lib/apt/lists/*

RUN go build -a -installsuffix cgo -o main ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /app/main /app/main
COPY --from=builder /app/config.yaml /app/config.yaml

EXPOSE 8080

ENTRYPOINT ["/app/main"]
