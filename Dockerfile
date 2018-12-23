FROM golang:1.11.2 as builder
WORKDIR /go/src/hlc
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

#dockerize app------------------------------------------------------------------
FROM alpine:3.8
RUN apk --no-cache add ca-certificates
#install mongo
RUN apk add --no-cache mongodb=3.6.7-r0
VOLUME ["/data/db", "tmp/data"]

COPY run.sh /root
ENTRYPOINT ["/root/run.sh"]

WORKDIR /root/
COPY --from=builder /go/src/hlc/app/app .
EXPOSE 80
CMD ["sh", "-c", "(mongod --bind_ip 0.0.0.0 &) && ./app"]