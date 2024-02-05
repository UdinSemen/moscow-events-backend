FROM golang:1.21-alpine AS builder

ARG SERVER_PORT=8080
WORKDIR /usr/local/src

RUN apk --no-cache add bash make gcc gettext musl-dev

# dependicies
COPY ["app/go.mod", "app/go.sum", "./"]
RUN go mod download

# build
COPY app ./
RUN go build -o ./bin/app cmd/rest-api/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/bin/app /
COPY config/dev.yaml /dev.yaml

# listen on 8080
EXPOSE ${SERVER_PORT}
CMD ["/app"]