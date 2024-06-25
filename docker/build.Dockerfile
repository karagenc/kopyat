FROM golang:alpine AS build

RUN apk --no-cache add \
	make \
	ca-certificates tzdata \
	&& update-ca-certificates

WORKDIR /src
COPY . .
RUN make

FROM alpine AS deploy

RUN apk --no-cache add \
	restic \
	ca-certificates tzdata \
	&& update-ca-certificates

COPY --from=build /src/kopyaship /usr/local/bin

ENTRYPOINT [ "kopyaship", "service" ]
