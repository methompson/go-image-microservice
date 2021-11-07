FROM alpine:3.14
# FROM golang:1.16

RUN mkdir -p /srv/image
WORKDIR /srv/image

COPY bin/image-microservice ./image-microservice
COPY firebase.json ./firebase.json

CMD ["./image-microservice"]