## Build
FROM golang:1.20-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /git-visualizer

## Deploy
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /git-visualizer /git-visualizer

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/git-visualizer"]