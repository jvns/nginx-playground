FROM golang:1.17 AS go

RUN git clone --depth=1 https://github.com/jvns/go-httpbin /go-httpbin
WORKDIR  /go-httpbin
RUN make

ADD ./api /app
WORKDIR /app
RUN go build
RUN go build ./cmd/run_nginx

FROM nginx:1.21

RUN apt-get update && apt-get install -y curl httpie && apt-get clean
RUN apt-get -y install libcap2-bin
RUN apt-get -y install bubblewrap
COPY --from=go /go-httpbin/dist/go-httpbin /usr/bin/go-httpbin
COPY --from=go /app/nginx-playground /app/nginx-playground
COPY --from=go /app/run_nginx /app/run_nginx

WORKDIR /app

CMD ["/app/nginx-playground"]
