# syntax=docker/dockerfile:1
FROM node:24-alpine AS frontend
WORKDIR /frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run check && npm test && npm run build

FROM golang:1.27.1-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/dota-leagues .

FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata && addgroup -g 10001 app && adduser -D -u 10001 -G app app
WORKDIR /app
COPY --from=backend /out/dota-leagues /app/dota-leagues
COPY --from=frontend /frontend/dist/ /app/public/
COPY config.json.example /app/config.json
RUN mkdir /data && chown app:app /data
USER app
ENV ASSET_DIR=/data
EXPOSE 1323
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 CMD wget -q -O /dev/null http://127.0.0.1:1323/healthz || exit 1
ENTRYPOINT ["/app/dota-leagues"]
CMD ["-production"]
