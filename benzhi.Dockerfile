FROM golang:1.26.2
ENV GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
WORKDIR /src
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
RUN go build -mod=vendor -o /mineair ./cmd/mineair
EXPOSE 21206
CMD ["/mineair"]
