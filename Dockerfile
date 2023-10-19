FROM alpine:3.18.4

RUN apk add --update ca-certificates \
    && rm -rf /var/cache/apk/*

ADD ./node-operator /node-operator

ENTRYPOINT ["/node-operator"]
