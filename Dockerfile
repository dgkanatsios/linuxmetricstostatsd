#build stage
FROM golang:1.13.1-alpine3.10 AS builder
RUN apk add --no-cache git
WORKDIR /usr/app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o ./build/linuxmetricstostatsd .

#final stage
FROM alpine:3.10
RUN apk --no-cache add ca-certificates
WORKDIR /usr/app
COPY --from=builder /usr/app/build .
CMD ["./linuxmetricstostatsd"]