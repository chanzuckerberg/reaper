FROM golang:1 as builder

WORKDIR /go/src/github.com/chanzuckerberg/reaper/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o reaper .

# app
FROM alpine:latest

RUN apk update && apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=0 /go/src/github.com/chanzuckerberg/reaper/reaper .

CMD ["./reaper"]
