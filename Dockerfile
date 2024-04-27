FROM golang:1.21 as builder
WORKDIR /app
COPY cmd cmd
COPY pkg pkg
COPY go.mod go.sum .
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags "-linkmode external -extldflags -static" -o tribler_arr_shim ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY scripts scripts
COPY --from=builder /app/tribler_arr_shim .
CMD ["./tribler_arr_shim", "server"]
