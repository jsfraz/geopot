# Builder stage for backend
FROM golang:1.26.10-alpine AS backend-build

WORKDIR /app

COPY geopot/ .

RUN go mod download

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o geopot

# Builder stage for frontend
FROM node:24-alpine AS frontend-build

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./

RUN npm ci --no-audit --no-fund

COPY frontend/ ./

RUN npm run build

# Final stage
FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 app \
    && adduser -S -D -H -u 10001 -G app app

WORKDIR /app

# Never default to debug mode in a container: it exposes Swagger/OpenAPI and
# logs SQL statements including captured credentials.
ENV GIN_MODE=release

COPY --from=backend-build /app/geopot .

COPY --from=backend-build /app/html ./html

COPY --from=frontend-build /app/dist ./static

RUN chown -R app:app /app

USER app

EXPOSE 8080

EXPOSE 2222

CMD ["./geopot"]
