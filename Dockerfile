# --- Stage 1: Build Frontend ---
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# --- Stage 2: Build Server ---
FROM golang:1.21-alpine AS server-builder
WORKDIR /app
# Install build dependencies
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy pre-built frontend from stage 1
COPY --from=frontend-builder /app/web/dist ./web/dist
RUN CGO_ENABLED=1 GOOS=linux go build -o colink-server ./cmd/server

# --- Stage 3: Final Image ---
FROM alpine:latest
RUN apk add --no-cache ca-certificates libc6-compat
WORKDIR /app
COPY --from=server-builder /app/colink-server .

EXPOSE 8080
CMD ["./colink-server"]
