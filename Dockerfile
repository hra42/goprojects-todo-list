FROM scratch

COPY tasks /app

ENTRYPOINT ["/app"]
