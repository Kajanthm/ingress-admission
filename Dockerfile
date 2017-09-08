FROM alpine:3.6
MAINTAINER Rohith Jayawardene <gambol99@gmail.com>

ADD bin/ingress-admission /ingress-admission

ENTRYPOINT [ "/ingress-admission" ]
