FROM golang:1.6
COPY . /go/src/github.com/kyleterry/sufr
WORKDIR /go/src/github.com/kyleterry/sufr
RUN make
RUN go install
EXPOSE 8090
CMD ["sufr"]
