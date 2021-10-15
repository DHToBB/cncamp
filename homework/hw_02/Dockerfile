FROM golang:1.17.2-alpine3.14 AS build

WORKDIR /go/src/httpserver/

COPY httpserver/* /go/src/httpserver/
RUN go env -w GO111MODULE=auto && go build -o /bin/httpserver

FROM alpine:3.13.6
COPY --from=build /bin/httpserver /bin/
EXPOSE 8080
ENTRYPOINT ["/bin/httpserver"]
