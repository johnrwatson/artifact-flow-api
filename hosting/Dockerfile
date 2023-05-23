# docker build --platform linux/amd64 -f ./hosting/Dockerfile -t artifact-flow-api:latest .
FROM golang:1.19-buster as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -v -o /artifact-flow-api

FROM debian:buster-slim

# Copy the binary to the production image from the builder stage.
COPY --from=builder /artifact-flow-api /artifact-flow-api

EXPOSE 8000

CMD [ "/artifact-flow-api" ]