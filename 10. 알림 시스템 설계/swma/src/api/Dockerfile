FROM golang:1.22-alpine

WORKDIR /app

# 모듈 의존성 해결을 위해 필요한 폴더 복사
COPY protocol ./protocol
COPY api ./api

WORKDIR /app/api
RUN go mod download
RUN go build -o api-server .

CMD ["./api-server"]
