FROM node:22 AS frontend
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY server/go.mod server/go.sum ./server/
COPY server/main.go ./server/
RUN cd server && CGO_ENABLED=0 go build -o bakery-server .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=frontend /app/dist /app/dist
COPY --from=backend /app/server/bakery-server /app/server/bakery-server
EXPOSE 3001
WORKDIR /app
CMD ["/app/server/bakery-server"]
