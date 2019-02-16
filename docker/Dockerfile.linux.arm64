FROM plugins/base:linux-arm64

LABEL maintainer="Bo-Yi Wu <appleboy.tw@gmail.com>" \
  org.label-schema.name="Drone telegram" \
  org.label-schema.vendor="Bo-Yi Wu" \
  org.label-schema.schema-version="1.0"

ADD release/linux/arm64/drone-telegram /bin/

ENTRYPOINT ["/bin/drone-telegram"]
