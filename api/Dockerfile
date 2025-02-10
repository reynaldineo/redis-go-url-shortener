FROM golang:alpine AS builder

RUN mkdir /build

ADD . /build/

WORKDIR /build

RUN go mod download
RUN go build -o main .

# stage 2

FROM alpine

RUN adduser -S -D -H -h /app appuser

USER appuser

COPY . /app

copy --from=builder /build/main /app/

WORKDIR /app

EXPOSE 8080

CMD ["./main"]
