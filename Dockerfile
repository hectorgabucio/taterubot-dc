# syntax=docker/dockerfile:1

## Build
FROM golang:1.18 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /taterubot
#RUN cp ./config.json /config.json

## Deploy
FROM debian:latest

WORKDIR /

RUN apt update && apt install ffmpeg ca-certificates -y

COPY --from=build /taterubot /taterubot
COPY --from=build /app/config.json /config.json
COPY --from=build /app/infrastructure/sql/migrations/* /infrastructure/sql/migrations/

ARG CLOUDAMQP_URL
ARG CLOUDAMQP_APIKEY
ARG DATABASE_URL
ARG LANGUAGE
ARG BASE_PATH
ARG BOT_TOKEN

ENTRYPOINT ["/taterubot"]