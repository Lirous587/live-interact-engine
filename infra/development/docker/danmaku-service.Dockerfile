FROM alpine:latest

WORKDIR /app

COPY build/danmaku-service .

EXPOSE 9093

CMD ["./danmaku-service"]
