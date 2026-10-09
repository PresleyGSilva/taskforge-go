FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api

FROM alpine:3.20
WORKDIR /app
COPY --from=build /api /app/api
COPY migrations /app/migrations
COPY openapi.yaml /app/openapi.yaml
EXPOSE 8083
ENV PORT=8083
CMD ["/app/api"]
