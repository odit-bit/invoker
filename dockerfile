########################################################
# STEP 1 use a temporary image to build a static binary
########################################################
FROM golang:1.21-alpine as build-stage

RUN apk add --no-cache git
RUN apk --no-cache add ca-certificates

WORKDIR /

COPY . .
RUN go mod download

# make static image
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o /monolith ./app/monolith/


########################################################
# STEP 2 create distroless image with trusted certs
########################################################
FROM gcr.io/distroless/base-debian11 AS build-release-stage
# RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /

COPY --from=build-stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-stage /monolith /monolith

EXPOSE 8080

ENTRYPOINT [ "./monolith" ]
# CMD [ "./monolith" ]