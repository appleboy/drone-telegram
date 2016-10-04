FROM centurylink/ca-certs

ADD drone-telegram /

ENTRYPOINT ["/drone-telegram"]
