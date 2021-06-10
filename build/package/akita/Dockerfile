FROM golang:latest

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

WORKDIR /go/src/
COPY . .
EXPOSE 8099
CMD ["/bin/bash", "/go/src/script/build.sh"]
