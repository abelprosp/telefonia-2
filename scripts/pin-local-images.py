#!/usr/bin/env python3
"""Pin deployment dependencies to verified, cached container image IDs."""
import datetime
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile


def inspect(kind, reference):
    return json.loads(subprocess.check_output(
        ["docker", kind, "inspect", reference], text=True))[0]


def main():
    root = Path(__file__).resolve().parent.parent
    env_path = root / ".env"
    original = env_path.read_bytes()
    pins = {}
    for variable, container in (
        ("POSTGRES_IMAGE", "postgres.connect.luxus"),
        ("RABBITMQ_IMAGE", "rabbitmq.connect.luxus"),
        ("MINIO_IMAGE", "minio.connect.luxus"),
        ("MINIO_MC_IMAGE", "minio-init.connect.luxus"),
    ):
        image_id = inspect("container", container)["Image"]
        if not re.fullmatch(r"sha256:[a-f0-9]{64}", image_id):
            raise RuntimeError(f"ID inválido para {container}")
        if inspect("image", image_id)["Id"] != image_id:
            raise RuntimeError(f"Imagem ausente para {container}")
        pins[variable] = image_id

    # Validate everything before changing .env. Preserve credentials and settings.
    contents = original.decode("utf-8")
    for variable, image_id in pins.items():
        pattern = rf"(?m)^(?:export\s+)?{variable}\s*=.*$"
        entry = f"{variable}={image_id}"
        if re.search(pattern, contents):
            contents = re.sub(pattern, entry, contents)
        else:
            contents = contents.rstrip("\r\n") + "\n" + entry + "\n"
    stamp = datetime.datetime.now().strftime("%Y%m%d-%H%M%S-%f")
    backup = root / f".env.backup.{stamp}"
    with backup.open("xb") as handle:
        os.chmod(backup, 0o600)
        handle.write(original)
    temporary = None
    try:
        with tempfile.NamedTemporaryFile(dir=root, delete=False) as handle:
            temporary = Path(handle.name)
            os.chmod(temporary, 0o600)
            handle.write(contents.encode("utf-8"))
        os.replace(temporary, env_path)
    finally:
        if temporary is not None and temporary.exists():
            temporary.unlink()
    print("Imagens locais verificadas e configuradas. Demais variáveis preservadas.")
    print(f"Backup: {backup.name}")


if __name__ == "__main__":
    main()
