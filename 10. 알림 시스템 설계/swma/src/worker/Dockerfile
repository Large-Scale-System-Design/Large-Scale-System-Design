FROM golang:1.22-alpine

WORKDIR /app

# 모듈 의존성 해결
COPY protocol ./protocol
COPY worker ./worker

WORKDIR /app/worker
RUN go mod download
RUN go build -o worker .

CMD ["./worker"]
