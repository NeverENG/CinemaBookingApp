FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/lterm ./cmd/lterm

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/lterm /usr/local/bin/lterm
COPY sql /app/sql
EXPOSE 8080
ENTRYPOINT ["/bin/sh", "-c"]
CMD ["lterm -migrate && exec lterm"]
