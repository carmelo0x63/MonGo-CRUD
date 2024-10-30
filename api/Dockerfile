FROM golang:1.23-alpine
WORKDIR /app
COPY go.* ./
RUN go mod download
#COPY . .
COPY main.go .
RUN go build -o main
EXPOSE 8080
CMD ["./main"]
