FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ayet-hadis-bot ./cmd/ayet-hadis-bot

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/ayet-hadis-bot /ayet-hadis-bot
ENTRYPOINT ["/ayet-hadis-bot"]
