FROM golang:1.20-bullseye
COPY ./go-azure ./bin
CMD ["./bin/go-azure"]