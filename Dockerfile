FROM golang:alpine as builder
WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 go build -o octopus-consumption-exporter

FROM alpine:latest  

COPY --from=builder /app/octopus-consumption-exporter /usr/local/bin

ENTRYPOINT ["./usr/local/bin/octopus-consumption-exporter"]