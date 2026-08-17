#!/usr/bin/env bash
set -e

echo "=== Configurando Nginx na VPS para telefonia.redobrai.online ==="

# 1. Copia a configuração atualizada para o sites-available
if [ -d "/etc/nginx/sites-available" ]; then
    cp docker/nginx/telefonia.redobrai.online.conf /etc/nginx/sites-available/telefonia.redobrai.online
    ln -sf /etc/nginx/sites-available/telefonia.redobrai.online /etc/nginx/sites-enabled/telefonia.redobrai.online
    # Remove default se houver conflito
    rm -f /etc/nginx/sites-enabled/default 2>/dev/null || true
elif [ -d "/etc/nginx/conf.d" ]; then
    cp docker/nginx/telefonia.redobrai.online.conf /etc/nginx/conf.d/telefonia.redobrai.online.conf
fi

echo "=== Testando configuração do Nginx ==="
nginx -t

echo "=== Recarregando Nginx ==="
systemctl reload nginx

echo "=== Testando resposta do Keycloak via Nginx ==="
curl -I http://127.0.0.1:8081/realms/luxus/.well-known/openid-configuration || echo "Keycloak container ainda iniciando na 8081..."

echo "=== Concluído com sucesso! ==="
