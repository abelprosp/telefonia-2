# Build SPA (Vite); variáveis VITE_* são embutidas no bundle nesta etapa.
FROM node:22-alpine AS build
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

ARG VITE_API_URL
ARG VITE_AUTH_URL
ARG VITE_CLIENT_ID
ARG VITE_CLIENT_SECRET
ARG VITE_STORAGE_BUCKET_NAME
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_AUTH_URL=$VITE_AUTH_URL
ENV VITE_CLIENT_ID=$VITE_CLIENT_ID
ENV VITE_CLIENT_SECRET=$VITE_CLIENT_SECRET
ENV VITE_STORAGE_BUCKET_NAME=$VITE_STORAGE_BUCKET_NAME

RUN npm run build

# Servir estáticos com fallback para SPA (PORT dinâmico para Railway/Heroku)
FROM nginx:1.27-alpine AS final
RUN apk add --no-cache gettext
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf.template /etc/nginx/conf.d/default.conf.template
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
ENV PORT=80
EXPOSE 80
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
