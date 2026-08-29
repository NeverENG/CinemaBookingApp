FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/lterm ./cmd/lterm

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S lterm && adduser -S -G lterm lterm
WORKDIR /app
COPY --from=build --chown=lterm:lterm /out/lterm /app/lterm
COPY --chown=lterm:lterm sql /app/sql
ENV TZ=Asia/Shanghai
USER lterm
EXPOSE 8080
ENTRYPOINT ["/app/lterm"]
