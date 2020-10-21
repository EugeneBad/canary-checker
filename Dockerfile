FROM flanksource/build-tools:v0.10.7 as builder
WORKDIR /app
COPY ./ ./
ARG NAME
ARG VERSION
RUN make static
RUN GOOS=linux GOARCH=amd64 go build -o canary-checker -ldflags "-X \"main.version=$VERSION\""  main.go

FROM golang:1.13.6
COPY --from=builder /app/canary-checker /app/
COPY --from=builder /app/fixtures /app/
WORKDIR /app
ENTRYPOINT ["/app/canary-checker"]
