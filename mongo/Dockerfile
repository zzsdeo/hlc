FROM alpine:3.8

RUN apk add --no-cache mongodb=3.6.7-r0

VOLUME /data/db

COPY run.sh /root
ENTRYPOINT ["/root/run.sh"]
CMD ["mongod", "--bind_ip", "0.0.0.0"]