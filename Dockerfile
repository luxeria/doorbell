FROM golang:1.11-alpine AS builder
WORKDIR /go/src/github.com/luxeria/doorbell
COPY . .
RUN CGO_ENABLED=0 go install -v ./...

FROM alpine:3.9
RUN apk add --no-cache ca-certificates mpg123
WORKDIR /doorbell
COPY --from=builder /go/src/github.com/luxeria/doorbell/assets ./assets
COPY --from=builder /go/bin/doorbell .
CMD ["./doorbell"]
