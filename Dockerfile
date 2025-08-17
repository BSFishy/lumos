FROM golang:1.24.4 AS build

WORKDIR /usr/src/app

# predownload dependencies
COPY go.mod go.sum ./
RUN go mod download

# build the application
COPY . .
RUN CGO_ENABLED=0 go build -v -o /lumos .

# run in distroless container
FROM gcr.io/distroless/static-debian12@sha256:2e114d20aa6371fd271f854aa3d6b2b7d2e70e797bb3ea44fb677afec60db22c
COPY --from=build /lumos /
CMD ["/lumos"]
