# Build SPA (Vite); variáveis VITE_* são embutidas no bundle nesta etapa.
FROM node:22-alpine AS build
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Railway: marque VITE_* como disponíveis no build (Settings → Variables).
# Não use ARG aqui — no Docker, ARG vazio sobrescreve variáveis injetadas pelo Railway.
RUN if [ -z "$VITE_AUTH_URL" ] || [ -z "$VITE_API_URL" ]; then \
      echo "ERRO: defina VITE_AUTH_URL e VITE_API_URL nas variáveis de build do Railway (connect-web)"; \
      exit 1; \
    fi

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
