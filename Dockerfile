FROM alpine:3.20.3

RUN apk add --update ca-certificates \
    && rm -rf /var/cache/apk/*

ADD ./node-operator /node-operator

ENTRYPOINT ["/node-operator"]
