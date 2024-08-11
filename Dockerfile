FROM scratch

COPY goprojects-todo-list /app

ENTRYPOINT ["/app"]
