FROM golang:1.19-buster as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -v -o /artifact-flow-api ./cmd/server

FROM debian:buster-slim

COPY --from=builder /artifact-flow-api /artifact-flow-api

EXPOSE 80

CMD [ "/artifact-flow-api" ]