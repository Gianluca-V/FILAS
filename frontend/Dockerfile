# Multi-stage build: compile the Vue SPA with Node, then serve the static
# dist/ with nginx (which also reverse-proxies /api to the Go api container).
# Used only by the docker-compose `demo` profile (see docker-compose.yml).
FROM node:20-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
