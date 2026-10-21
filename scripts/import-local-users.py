#!/usr/bin/env python3
"""Import an offline user export without credentials for an identity provider."""
import argparse
import json
from pathlib import Path
import subprocess
import sys


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("file", type=Path, help="JSON de usuários exportados")
    parser.add_argument("--assign-org", action="append", default=[], metavar="USUARIO=UUID")
    parser.add_argument("--assign-unassigned-to", metavar="UUID")
    args = parser.parse_args()
    if args.file.stat().st_size > 16 * 1024 * 1024:
        raise ValueError("Arquivo de usuários excede o limite de 16 MiB.")
    raw = args.file.read_bytes()
    data = json.loads(raw)
    users = data.get("users") if isinstance(data, dict) else data
    if not isinstance(users, list) or not all(isinstance(user, dict) for user in users):
        raise ValueError("Informe uma lista de usuários ou um objeto com users: [].")
    root = Path(__file__).resolve().parent.parent
    cmd = ["docker", "compose", "-f", "docker-compose.yml", "-f", "docker-compose.prod.yml",
           "run", "--rm", "--no-deps", "-T", "--pull", "never", "connect-api", "auth", "import-users", "-"]
    for assignment in args.assign_org:
        cmd.extend(["--assign-org", assignment])
    if args.assign_unassigned_to:
        cmd.extend(["--assign-unassigned-to", args.assign_unassigned_to])
    subprocess.run(cmd, input=raw, cwd=root, check=True)
    print("Importação concluída. Contas locais existentes foram preservadas.")


if __name__ == "__main__":
    try:
        main()
    except subprocess.CalledProcessError:
        sys.exit("Importação interrompida. Corrija o erro informado pela API e repita.")
    except (OSError, ValueError):
        sys.exit("Não foi possível importar: verifique se o arquivo JSON existe, está válido e tem até 16 MiB.")
