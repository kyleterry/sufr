FROM golang:1.10-alpine as builder
WORKDIR /go/src/github.com/kyleterry/sufr
COPY . .
RUN apk --no-cache add make git
RUN make

FROM alpine:3.4
RUN apk --no-cache add bash
COPY --from=builder /go/src/github.com/kyleterry/sufr/bin/sufr /usr/bin/sufr
VOLUME /var/lib/sufr
EXPOSE 8090
CMD ["sufr", "-bind", ":8090", "-data-dir", "/var/lib/sufr"]
