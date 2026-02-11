FROM golang:1.25.7-alpine AS builder

RUN apk add --no-cache git

WORKDIR /Orphie

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /Orphie/bin/app .


FROM alpine:latest

RUN apk add --no-cache ca-certificates 

COPY --from=builder /Orphie/bin/app /app

EXPOSE 307

ENTRYPOINT [ "/app" ]