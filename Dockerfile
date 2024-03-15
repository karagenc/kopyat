FROM golang:alpine AS build

RUN apk --no-cache add \
	make \
	ca-certificates tzdata \
	&& update-ca-certificates

WORKDIR /src
COPY . .
RUN make

FROM alpine AS deploy

COPY --from=build /src/kopyaship /usr/local/bin
COPY --from=build /src/kopyashipd /usr/local/bin

ENTRYPOINT [ "kopyashipd" ]
