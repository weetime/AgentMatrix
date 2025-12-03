ARG ARCH=amd64

FROM quanzhenglong.com/camp/golang:1.23.rc3-${ARCH} AS build

WORKDIR /go/src/app

COPY . .
RUN go mod tidy && make generate && make linux

FROM quanzhenglong.com/camp/alpine:3.18.6.rc1-${ARCH}
RUN mkdir -p /apps/logs
COPY --from=build /go/src/app/build/ /apps/
COPY --from=build /go/src/app/config/ /apps/config/