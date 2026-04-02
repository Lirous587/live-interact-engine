FROM alpine:latest

WORKDIR /app

COPY build/api-service .

COPY .env.* ./

# 暴露端口：8080（API）和 9100（Metrics）
EXPOSE 8080 9100

CMD ["./api-service"]
