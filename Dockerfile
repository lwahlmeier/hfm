FROM alpine:3.9

RUN apk update && \
    apk -Uuv add dumb-init ca-certificates bash && \
    rm /var/cache/apk/*
COPY run.sh /run.sh
RUN touch env.sh && chmod 755 /run.sh
COPY bin/hfm /hfm

EXPOSE 8844/tcp

ENTRYPOINT ["/run.sh"]
CMD ["/hfm"]

