FROM golang:1.13
WORKDIR /app
COPY . /app
ENV GOPROXY=https://goproxy.io
RUN make build


FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /rendora
COPY --from=0 /app /usr/bin
ENTRYPOINT ["rendora", "start", "-c", "/etc/rendora/config.yml"]