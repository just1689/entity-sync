# Build Golang app in Go container
FROM golang:1.12.4-alpine3.9 as builder
RUN mkdir /user && \
    echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
    echo 'nobody:x:65534:' > /user/group
RUN apk add --no-cache ca-certificates git
WORKDIR /src
COPY ../go.mod /src/
RUN go mod download
COPY .. .
WORKDIR /src/example
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app example.go

# Final app
FROM scratch as final
COPY --from=builder /user/group /user/passwd /etc/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/app /app
USER nobody:nobody
ENTRYPOINT ["/app"]