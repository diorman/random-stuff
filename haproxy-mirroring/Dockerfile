FROM golang:1.17-alpine AS go-builder

WORKDIR /src
COPY ./go.mod ./go.sum /src/
RUN go mod download

COPY ./backend/main.go /src/backend/main.go
RUN CGO_ENABLED=0 go build -o /src/bin/backend /src/backend/main.go

COPY ./spoa/main.go /src/spoa/main.go
RUN CGO_ENABLED=0 go build -o /src/bin/spoa /src/spoa/main.go

COPY ./dest/main.go /src/dest/main.go
RUN CGO_ENABLED=0 go build -o /src/bin/dest /src/dest/main.go

# -----------------------------------------------------

FROM alpine:3.16
COPY --from=go-builder /src/bin/backend /usr/bin/backend
COPY --from=go-builder /src/bin/spoa /usr/bin/spoa
COPY --from=go-builder /src/bin/dest /usr/bin/dest
