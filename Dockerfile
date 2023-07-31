FROM golang:1.20-alpine as build

WORKDIR /app

COPY . .

RUN go build -o /ktemplate .

FROM alpine

COPY --from=build /ktemplate /usr/local/bin/

ENTRYPOINT /usr/local/bin/ktemplate
