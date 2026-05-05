FROM node:22-alpine AS web-builder
WORKDIR /app/web
COPY web/package.json web/bun.lock ./
RUN npm install -g bun && bun install
COPY web .
RUN bun run build

FROM golang:1.25-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG DEPLOY_GIT_SHA=unknown
RUN echo "$DEPLOY_GIT_SHA" >/tmp/.leetbot-deploy-sha
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot ./cmd/bot
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o backfill ./cmd/backfill-analytics

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=go-builder /app/bot .
COPY --from=go-builder /app/server .
COPY --from=go-builder /app/backfill .
COPY --from=web-builder /app/web/dist ./web/dist
COPY start-all.sh .
RUN chmod +x start-all.sh

EXPOSE 8080
CMD ["./start-all.sh"]