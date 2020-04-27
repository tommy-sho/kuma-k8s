FROM golang:1.14-alpine as builder
RUN apk add --no-cache ca-certificates git
ENV PROJECT /github.com/tommy-sho/kuma-k8s/backend
WORKDIR /go/src/$PROJECT

ENV GO111MODULE on
COPY . .
RUN go install .

FROM alpine as release
RUN apk add --no-cache ca-certificates git
COPY --from=builder /go/bin/backend /backend
ENTRYPOINT ["/backend"]
