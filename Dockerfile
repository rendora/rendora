FROM golang:1.21.0
WORKDIR /app
COPY . /app
RUN make build


FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /rendora
COPY --from=0 /app/cmd/rendora/rendora /usr/bin
ENTRYPOINT ["rendora"]
