# Build
FROM golang:latest AS builder
LABEL maintainer="Ruchida Pithaksiripan <rpithaksiripan@gmail.com>"
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o k-tax .

# Deploy
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/k-tax .
EXPOSE 8080
ENV PORT=8080
CMD ["./k-tax"]