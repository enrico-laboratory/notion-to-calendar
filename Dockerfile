FROM golang:1.18
# Add the module files and download dependencies.
ENV GO111MODULE=on

COPY ./go.mod /go/src/app/go.mod
#COPY ./go.sum /go/src/app/go.sum
COPY ./cmd/app /go/src/app/
COPY ./vendor /go/src/app/vendor
WORKDIR /go/src/app

RUN env GOARCH=amd64 go build -o /app .

CMD [ "/app" ]
