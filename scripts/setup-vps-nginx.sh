#!/usr/bin/env bash
set -e

echo "=== Configurando Nginx na VPS para telefonia.redobrai.online ==="

if [ -d "/etc/nginx/sites-available" ]; then
    cp docker/nginx/telefonia.redobrai.online.conf /etc/nginx/sites-available/telefonia.redobrai.online
    ln -sf /etc/nginx/sites-available/telefonia.redobrai.online /etc/nginx/sites-enabled/telefonia.redobrai.online
    rm -f /etc/nginx/sites-enabled/default 2>/dev/null || true
elif [ -d "/etc/nginx/conf.d" ]; then
    cp docker/nginx/telefonia.redobrai.online.conf /etc/nginx/conf.d/telefonia.redobrai.online.conf
fi

echo "=== Testando configuração do Nginx ==="
nginx -t

echo "=== Recarregando Nginx ==="
systemctl reload nginx

echo "=== Verificando API do Telefonia ==="
curl -fsS "http://127.0.0.1:8010/health"
echo
echo "=== Concluído. Login próprio disponível em /api/v1/auth/. ==="
