# Air Social

## I. Start Project

### Hot Reload (Dev)
Requires Dockerfile:

```dockerfile
RUN go install github.com/air-verse/air@latest
CMD ["air"]
```

Run with Air + volume mount:

```sh
make docker-reload-mount
```

**Behavior**
- Host source code is mounted into the container.
- Air rebuilds + restarts automatically on file changes.
- No image rebuild.
- No container recreate.


### No Hot Reload (Binary Mode)
Requires Dockerfile:

```dockerfile
RUN go build -o server .
CMD ["./server"]
```

Run:

```sh
make docker-reload
```

**Behavior**
- Code inside the container is fixed at build time (binary baked into the image).
- Local code changes do NOT affect the running container.
- To update code you must: **Rebuild the image** and **Recreate the container**


### Summary
- **Mount (Dev mode)** → instant reload via Air, no rebuilds, no container recreate.  
- **No mount (Binary mode)** → production-style, must rebuild image + recreate container when code changes.
