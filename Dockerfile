FROM golang:1.11
WORKDIR /app
COPY ./rendora /app
RUN CGO_ENABLED=0 go build


FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /rendora
COPY --from=0 /app/rendora .
ENTRYPOINT ["/rendora/rendora"]