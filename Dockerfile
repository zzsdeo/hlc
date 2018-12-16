FROM golang:1.11.2 as builder
WORKDIR /go/src/ProjectDB
#copy sources
COPY app ./app
COPY Gopkg.lock Gopkg.toml ./
#install dep
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
#install app dependencies
RUN dep ensure
#build app
WORKDIR app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

#dockerize app
FROM alpine:3.8
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/ProjectDB/app/app .
CMD ["./app"]