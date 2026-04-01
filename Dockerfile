FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY internal/ ./internal/

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /server .

FROM backplane/upx:latest AS packer

COPY --from=builder /server /tmp/server
RUN upx --best --lzma -o /server /tmp/server

FROM gcr.io/distroless/static:nonroot

USER nonroot
WORKDIR /app
COPY --from=packer /server /app/server
COPY /static /app/static

EXPOSE 8080
CMD ["/app/server"]
