# Build from the repo root: docker build -f examples/demoserver/Dockerfile -t mctop-demo .
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /demoserver ./examples/demoserver

FROM gcr.io/distroless/static-debian12
COPY --from=build /demoserver /demoserver
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/demoserver"]
