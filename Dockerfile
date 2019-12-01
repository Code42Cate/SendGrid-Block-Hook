FROM golang:latest

LABEL maintainer="Jonas Scholz <info@jonas-scholz.me>"
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main
ENV INTERVAL=5
ENV DISCORD_TOKEN=""
ENV DISCORD_CHANNEL_ID=""
ENV SENDGRID_TOKEN=""
ENV LAST_TIMESTAMP=-1

EXPOSE 8080
ENTRYPOINT [ "./main"]