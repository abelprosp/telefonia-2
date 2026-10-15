#!/usr/bin/env python3
"""Generate a persistent MFA encryption key without printing credentials."""
from datetime import datetime
import os
from pathlib import Path
import re
import secrets


def main():
    path = Path(__file__).resolve().parent.parent / ".env"
    original = path.read_bytes() if path.exists() else b""
    text = original.decode("utf-8")
    match = re.search(r"(?m)^(?:export\s+)?AUTH_SECRET\s*=(.*)$", text)
    exported = os.environ.get("AUTH_SECRET", "")
    if exported and len(exported) < 32:
        raise SystemExit("AUTH_SECRET exportado precisa ter pelo menos 32 caracteres.")
    if match and match.group(1).strip().strip("\"'"):
        if exported and exported != match.group(1).strip().strip("\"'"):
            raise SystemExit("AUTH_SECRET exportado difere do .env. Preserve a chave utilizada na instalação.")
        if len(match.group(1).strip().strip("\"'")) < 32:
            raise SystemExit("AUTH_SECRET precisa ter pelo menos 32 caracteres.")
        print("Autenticação própria: chave existente preservada.")
        return
    if original:
        backup = path.with_name(".env.backup." + datetime.now().strftime("%Y%m%d-%H%M%S-%f"))
        with backup.open("xb") as handle:
            os.chmod(backup, 0o600)
            handle.write(original)
    entry = "AUTH_SECRET=" + (exported or secrets.token_hex(32))
    if match:
        text = re.sub(r"(?m)^(?:export\s+)?AUTH_SECRET\s*=.*$", entry, text)
    else:
        text = text.rstrip("\r\n") + "\n" + entry + "\n"
    # Exclusive temporary file, restricted before any secret is written.
    temporary = path.with_name(".env.auth-" + secrets.token_hex(8))
    try:
        with temporary.open("xb") as handle:
            os.chmod(temporary, 0o600)
            handle.write(text.encode("utf-8"))
        os.replace(temporary, path)
    finally:
        if temporary.exists():
            temporary.unlink()
    print("Autenticação própria configurada. Demais variáveis preservadas.")


if __name__ == "__main__":
    main()
