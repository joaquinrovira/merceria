FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /server .

FROM gcr.io/distroless/static:nonroot


USER nonroot

WORKDIR /app
COPY --from=builder /server /app/server
COPY /static /app/static

EXPOSE 8080
CMD ["/app/server"]
