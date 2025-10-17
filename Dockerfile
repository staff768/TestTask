FROM golang:1.25-alpine

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o main ./cmd/web

EXPOSE 8080

CMD ["./main"]
LABEL authors="staff"