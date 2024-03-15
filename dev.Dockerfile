FROM golang:alpine

RUN apk --update-cache add \
    zsh bash make git curl \
	tar ca-certificates tzdata \
    python3 \
    py3-pip
RUN update-ca-certificates

WORKDIR /src
RUN git config --global --add safe.directory /src
COPY . .
ENTRYPOINT [ "tail", "-f", "/dev/null" ]
