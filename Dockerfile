FROM golang:1 as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o reaper .

# Install chamber
ENV CHAMBER_VERSION=v2.8.2
RUN wget -q https://github.com//segmentio/chamber/releases/download/${CHAMBER_VERSION}/chamber-${CHAMBER_VERSION}-linux-amd64 -O /bin/chamber
RUN chmod +x /bin/chamber

# app
FROM alpine:latest

RUN apk update && apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/reaper /bin/reaper
COPY --from=builder /bin/chamber /bin/chamber

CMD ["reaper"]
