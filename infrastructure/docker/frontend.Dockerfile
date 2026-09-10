FROM node:22-alpine AS build

WORKDIR /app
COPY apps/frontend/package*.json ./
RUN npm ci
COPY apps/frontend/ ./
ARG VITE_API_URL=/api
RUN VITE_API_URL="$VITE_API_URL" npm run build

FROM nginx:1.27-alpine

COPY --from=build /app/dist /usr/share/nginx/html
COPY infrastructure/nginx/frontend.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
