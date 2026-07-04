import os, base64, hashlib

password: str | None = ""

while password == "" or password is None:
    password = input("Enter desired password: ")                              # set a strong password

iterations = int(input("Enter number of iterations: ")) or 27500              # Keycloak's current PBKDF2-SHA256 default

salt = os.urandom(16)
dk = hashlib.pbkdf2_hmac("sha256", password.encode("utf-8"), salt, iterations, dklen=64)

print("salt:", base64.b64encode(salt).decode())
print("hash:", base64.b64encode(dk).decode())
