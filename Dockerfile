FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /dozingo ./cmd/api

FROM alpine:3.21

RUN apk add --no-cache ca-certificates
COPY --from=build /dozingo /dozingo

EXPOSE 8080

ENTRYPOINT ["/dozingo"]
