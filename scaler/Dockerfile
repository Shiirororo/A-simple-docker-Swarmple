FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -mod=vendor -o scaler .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/scaler .
EXPOSE 3619
CMD ["./scaler"]
