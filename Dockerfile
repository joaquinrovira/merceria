FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

FROM gcr.io/distroless/static:nonroot

COPY --from=builder /server /server
COPY /static /static

EXPOSE 8080

USER nonroot

CMD ["/server"]
