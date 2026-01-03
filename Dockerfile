ARG APPVERSION="local"

FROM golang:1.24 AS build
WORKDIR /go/src/app
ENV GOCACHE=/go/cache
ENV CGO_ENABLED=0

COPY . .
RUN --mount=type=cache,target="/go/cache" \
  echo "Building version: $APPVERSION" && \
  go mod download && \
  go build -ldflags "-X cmd.controller.main.APPVERSION=$APPVERSION" -o /go/bin/emptynscontroller ./cmd/controller/

FROM gcr.io/distroless/static-debian12
CMD ["/app/emptynscontroller"]
COPY --from=build /go/bin/ /app
