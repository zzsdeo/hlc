FROM golang:1.11.2 as builder
WORKDIR /go/src/hlc
#install easyjson
#RUN go get -u github.com/mailru/easyjson/...
#copy sources
COPY app ./app
COPY Gopkg.lock Gopkg.toml ./
#generate easyjson marshallers/unmarshallers
#WORKDIR /go/src/hlc/app/models
#RUN easyjson -all account.go
#install dep
WORKDIR /go/src/hlc
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
#install app dependencies
RUN dep ensure
#build app
WORKDIR /go/src/hlc/app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

#dockerize app------------------------------------------------------------------
FROM alpine:3.8
RUN apk --no-cache add ca-certificates
VOLUME ["tmp/data"]
WORKDIR /root/
COPY --from=builder /go/src/hlc/app/app .
EXPOSE 80
CMD ["./app"]