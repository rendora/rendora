FROM golang:1.11
WORKDIR /app
COPY . /app
RUN make build


FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /rendora
COPY --from=0 /app/cmd/rendora/rendora /usr/bin
ENTRYPOINT ["rendora"]