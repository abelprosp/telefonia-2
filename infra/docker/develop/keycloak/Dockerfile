FROM quay.io/keycloak/keycloak:latest

ENV KC_DB=postgres \
    KC_PROXY_HEADERS=xforwarded \
    KC_HTTP_ENABLED=true \
    KC_HOSTNAME_STRICT=false \
    KC_HEALTH_ENABLED=true \
    KC_METRICS_ENABLED=true

# Optional: Pre-build the Keycloak configuration for faster startup
RUN /opt/keycloak/bin/kc.sh build \
    --health-enabled=true \
    --metrics-enabled=true

EXPOSE 8080

# Healthcheck - Keycloak exposes /health/ready and /health/live
HEALTHCHECK --interval=30s \
    --timeout=10s \
    --start-period=40s \
    --retries=5 \
    CMD curl -f http://localhost:8080/health/ready || exit 1

# Start in production mode
ENTRYPOINT ["/opt/keycloak/bin/kc.sh", "--verbose"]
