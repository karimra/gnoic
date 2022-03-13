FROM alpine

LABEL maintainer="Karim Radhouani <medkarimrdi@gmail.com>"
LABEL documentation="https://gnoic.kmrd.dev"
LABEL repo="https://github.com/karimra/gnoic"

COPY gnoic /app/gnoic
ENTRYPOINT [ "/app/gnoic" ]
CMD [ "help" ]
