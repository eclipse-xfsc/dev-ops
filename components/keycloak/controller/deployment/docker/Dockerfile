FROM golang:1.24-bookworm AS builder
WORKDIR /src
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -trimpath -o /controller .
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /controller /controller
ENTRYPOINT ["/controller"]
