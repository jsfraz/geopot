# Builder stage for backend
FROM golang:1.25.1-alpine AS backend-build

WORKDIR /app

COPY geopot/ .

RUN go mod download

RUN go build -o geopot

# Builder stage for frontend
FROM node:22-alpine AS frontend-build

WORKDIR /app

COPY frontend/ .

RUN npm install

RUN npm run build

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=backend-build /app/geopot .

COPY --from=frontend-build /app/dist ./static

EXPOSE 8080

EXPOSE 2222

CMD ["./geopot"]