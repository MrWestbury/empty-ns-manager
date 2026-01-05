ARG APPVERSION="local"

FROM golang:1.24 AS build
WORKDIR /go/src/app

COPY . .
RUN echo "Building version: $APPVERSION" && \
  go mod download && \
  go build -ldflags "-X main.APPVERSION=$APPVERSION" -o /go/bin/emptynscontroller ./cmd/controller/

FROM gcr.io/distroless/static-debian12
CMD ["/app/emptynscontroller"]
COPY --from=build /go/bin/ /app
