#!/usr/bin/env python3
"""One-time migration using the old container credentials, without printing them."""
import json
import os
from pathlib import Path
import subprocess


def main():
    root = Path(__file__).resolve().parent.parent
    record = json.loads(subprocess.check_output(
        ["docker", "container", "inspect", "keycloak.connect.luxus"], text=True))[0]
    old_env = dict(item.split("=", 1) for item in record["Config"]["Env"] if "=" in item)
    env = os.environ.copy()
    env["KEYCLOAK_AUTH_SERVER_URL"] = "http://keycloak.connect.luxus:8080/auth"
    env["KEYCLOAK_REALM"] = "luxus"
    env["KEYCLOAK_RESOURCE"] = "connect-cli"
    env["KEYCLOAK_ADMIN_USERNAME"] = env.get("KEYCLOAK_ADMIN_USERNAME") or old_env.get("KC_BOOTSTRAP_ADMIN_USERNAME") or old_env.get("KEYCLOAK_ADMIN") or "admin"
    env["KEYCLOAK_ADMIN_PASSWORD"] = env.get("KEYCLOAK_ADMIN_PASSWORD") or old_env.get("KC_BOOTSTRAP_ADMIN_PASSWORD") or old_env.get("KEYCLOAK_ADMIN_PASSWORD", "")
    if not env["KEYCLOAK_ADMIN_PASSWORD"]:
        raise SystemExit("Credencial antiga ausente. Defina KEYCLOAK_ADMIN_PASSWORD apenas para esta migração.")
    cmd = ["docker", "compose", "-f", "docker-compose.yml", "-f", "docker-compose.prod.yml",
           "run", "--rm", "--no-deps", "-T", "--pull", "never"]
    for key in ("KEYCLOAK_AUTH_SERVER_URL", "KEYCLOAK_REALM", "KEYCLOAK_RESOURCE", "KEYCLOAK_ADMIN_USERNAME", "KEYCLOAK_ADMIN_PASSWORD"):
        cmd.extend(["-e", key])
    cmd.extend(["connect-api", "auth", "import-keycloak"])
    subprocess.run(cmd, cwd=root, env=env, check=True)


if __name__ == "__main__":
    main()
