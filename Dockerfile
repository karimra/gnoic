FROM golang:1.19.5 as builder
ADD . /build
WORKDIR /build
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o gnoic .

FROM alpine

LABEL maintainer="Karim Radhouani <medkarimrdi@gmail.com>"
LABEL documentation="https://gnoic.kmrd.dev"
LABEL repo="https://github.com/karimra/gnoic"
COPY --from=builder /build/gnoic /app/
ENTRYPOINT [ "/app/gnoic" ]
CMD [ "help" ]
