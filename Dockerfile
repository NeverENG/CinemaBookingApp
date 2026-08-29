FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/lterm ./cmd/lterm

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/lterm /app/lterm
COPY sql /app/sql
EXPOSE 8080
ENTRYPOINT ["/app/lterm"]
