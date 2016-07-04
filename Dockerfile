FROM golang:1.6
MAINTAINER Kyle Terry "kyle@kyleterry.com"
COPY . /go/src/github.com/kyleterry/sufr
WORKDIR /go/src/github.com/kyleterry/sufr
RUN make
RUN cp sufr /go/bin
VOLUME ["/root/.config/sufr/data"]
EXPOSE 8090
CMD ["sufr", "-bind", ":8090"]
