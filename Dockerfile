# Multi-stage build: Go backend + React UI
FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mxclone main.go

FROM node:20 AS ui-builder
WORKDIR /ui
COPY ui/package.json ui/package-lock.json ./
RUN npm ci
COPY ui .
RUN npm run build

FROM alpine:3.19 AS final
WORKDIR /app
COPY --from=builder /app/mxclone ./mxclone
COPY --from=ui-builder /ui/dist ./ui/dist
COPY ui/public ./ui/public
EXPOSE 8080
ENV UI_DIST_PATH=/app/ui/dist
CMD ["./mxclone", "api"]
