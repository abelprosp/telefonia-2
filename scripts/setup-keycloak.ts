#!/usr/bin/env npx ts-node
/**
 * setup-keycloak.ts
 *
 * Importa o realm 'luxus' no Keycloak de forma idempotente.
 * Cria roles, client scopes, cliente e usuários de teste caso ainda não existam.
 *
 * Uso:
 *   npx ts-node scripts/setup-keycloak.ts
 *
 * Variáveis de ambiente (todas opcionais — os defaults apontam para produção):
 *   KEYCLOAK_URL           URL base do Keycloak  (default: https://keycloak-production-734c.up.railway.app)
 *   KEYCLOAK_ADMIN_USER    Usuário admin          (default: admin)
 *   KEYCLOAK_ADMIN_PASS    Senha admin            (default: admin)
 *   KEYCLOAK_REALM         Realm a importar       (default: luxus)
 */

const KEYCLOAK_URL =
  (process.env.KEYCLOAK_URL ?? "https://keycloak-production-734c.up.railway.app").replace(/\/$/, "");
const ADMIN_USER = process.env.KEYCLOAK_ADMIN_USER ?? "admin";
const ADMIN_PASS = process.env.KEYCLOAK_ADMIN_PASS ?? "admin";
const REALM = process.env.KEYCLOAK_REALM ?? "luxus";

// ---------------------------------------------------------------------------
// Tipos mínimos
// ---------------------------------------------------------------------------

interface TokenResponse {
  access_token: string;
  expires_in: number;
}

interface RealmRole {
  id: string;
  name: string;
}

// ---------------------------------------------------------------------------
// Definição do realm
// ---------------------------------------------------------------------------

const REALM_ROLES = [
  { name: "admin",     description: "Administrador Luxus Connect" },
  { name: "user",      description: "Utilizador padrão" },
  { name: "partner",   description: "Parceiro comercial com acesso restrito" },
  { name: "master",    description: "Acesso total ao sistema e gestão de usuários" },
  { name: "employee",  description: "Funcionário — operação básica sem módulo financeiro" },
  { name: "financial", description: "Financeiro — controle financeiro completo" },
];

const ORG_ATTRIBUTE =
  '{"luxus":{"id":"00000000-0000-0000-0000-000000000001","name":["Luxus Connect"]}}';

const TEST_USERS = [
  {
    username: "dev",
    password: "dev",
    firstName: "Dev",
    lastName: "Luxus",
    email: "dev@luxus.local",
    roles: ["master", "admin", "user"],
  },
  {
    username: "parceiro",
    password: "parceiro",
    firstName: "Parceiro",
    lastName: "Luxus",
    email: "parceiro@luxus.local",
    roles: ["partner", "user"],
  },
  {
    username: "funcionario",
    password: "funcionario",
    firstName: "Funcionario",
    lastName: "Luxus",
    email: "funcionario@luxus.local",
    roles: ["employee", "user"],
  },
  {
    username: "financeiro",
    password: "financeiro",
    firstName: "Financeiro",
    lastName: "Luxus",
    email: "financeiro@luxus.local",
    roles: ["financial", "user"],
  },
];

const CLIENT_SCOPES = [
  {
    name: "organization",
    description: "Claim organization no token",
    protocol: "openid-connect",
    attributes: {
      "include.in.token.scope": "true",
      "display.on.consent.screen": "false",
    },
    protocolMappers: [
      {
        name: "organization-mapper",
        protocol: "openid-connect",
        protocolMapper: "oidc-usermodel-attribute-mapper",
        consentRequired: false,
        config: {
          "user.attribute": "organization",
          "claim.name": "organization",
          "jsonType.label": "JSON",
          "id.token.claim": "true",
          "access.token.claim": "true",
          "userinfo.token.claim": "true",
          multivalued: "false",
        },
      },
    ],
  },
  {
    name: "luxus-roles",
    description: "Realm roles no access token",
    protocol: "openid-connect",
    attributes: {
      "include.in.token.scope": "true",
      "display.on.consent.screen": "false",
    },
    protocolMappers: [
      {
        name: "realm-roles",
        protocol: "openid-connect",
        protocolMapper: "oidc-usermodel-realm-role-mapper",
        consentRequired: false,
        config: {
          multivalued: "true",
          "userinfo.token.claim": "true",
          "id.token.claim": "true",
          "access.token.claim": "true",
          "claim.name": "roles",
          "jsonType.label": "String",
        },
      },
    ],
  },
];

const CONNECT_CLIENT = {
  clientId: "connect-cli",
  name: "Luxus Connect SPA",
  enabled: true,
  publicClient: true,
  standardFlowEnabled: true,
  directAccessGrantsEnabled: true,
  implicitFlowEnabled: false,
  serviceAccountsEnabled: false,
  redirectUris: [
    "http://localhost:5173/*",
    "http://localhost:3000/*",
    "http://localhost:8002/*",
    "http://127.0.0.1:5173/*",
  ],
  webOrigins: [
    "http://localhost:5173",
    "http://localhost:3000",
    "http://localhost:8002",
    "http://127.0.0.1:5173",
  ],
  protocol: "openid-connect",
  fullScopeAllowed: true,
  attributes: {
    "client.use.lightweight.access.token.enabled": "false",
  },
  defaultClientScopes: ["openid", "profile", "email", "organization", "luxus-roles"],
  optionalClientScopes: ["offline_access"],
  protocolMappers: [
    {
      name: "realm-roles-direct",
      protocol: "openid-connect",
      protocolMapper: "oidc-usermodel-realm-role-mapper",
      consentRequired: false,
      config: {
        multivalued: "true",
        "userinfo.token.claim": "true",
        "id.token.claim": "true",
        "access.token.claim": "true",
        "claim.name": "roles",
        "jsonType.label": "String",
      },
    },
  ],
};

// ---------------------------------------------------------------------------
// Helpers HTTP
// ---------------------------------------------------------------------------

async function getAdminToken(): Promise<string> {
  log("Obtendo token de acesso admin...");
  const body = new URLSearchParams({
    grant_type: "password",
    client_id: "admin-cli",
    username: ADMIN_USER,
    password: ADMIN_PASS,
  });

  const res = await fetch(`${KEYCLOAK_URL}/realms/master/protocol/openid-connect/token`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: body.toString(),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Falha ao obter token admin (${res.status}): ${text}`);
  }

  const data = (await res.json()) as TokenResponse;
  log("Token obtido com sucesso.");
  return data.access_token;
}

async function adminFetch(
  token: string,
  method: string,
  path: string,
  body?: unknown,
): Promise<Response> {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${token}`,
  };
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  return fetch(`${KEYCLOAK_URL}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
}

// ---------------------------------------------------------------------------
// Etapas de setup
// ---------------------------------------------------------------------------

async function ensureRealm(token: string): Promise<void> {
  log(`Verificando realm '${REALM}'...`);
  const res = await adminFetch(token, "GET", `/admin/realms/${REALM}`);

  if (res.status === 404) {
    log(`Realm '${REALM}' não encontrado. Criando...`);
    const createRes = await adminFetch(token, "POST", "/admin/realms", {
      realm: REALM,
      enabled: true,
      registrationAllowed: false,
      loginWithEmailAllowed: true,
      duplicateEmailsAllowed: false,
      resetPasswordAllowed: true,
      editUsernameAllowed: false,
      bruteForceProtected: false,
      defaultRoles: ["user"],
    });
    if (!createRes.ok) {
      const text = await createRes.text();
      throw new Error(`Falha ao criar realm (${createRes.status}): ${text}`);
    }
    log(`Realm '${REALM}' criado.`);
  } else if (res.ok) {
    log(`Realm '${REALM}' já existe.`);
  } else {
    const text = await res.text();
    throw new Error(`Erro ao verificar realm (${res.status}): ${text}`);
  }
}

async function ensureRealmRoles(token: string): Promise<void> {
  log("Configurando realm roles...");

  const listRes = await adminFetch(token, "GET", `/admin/realms/${REALM}/roles`);
  if (!listRes.ok) {
    throw new Error(`Falha ao listar roles (${listRes.status})`);
  }
  const existing = (await listRes.json()) as Array<{ name: string }>;
  const existingNames = new Set(existing.map((r) => r.name));

  for (const role of REALM_ROLES) {
    if (existingNames.has(role.name)) {
      log(`  Role '${role.name}' já existe. Pulando.`);
      continue;
    }
    const res = await adminFetch(token, "POST", `/admin/realms/${REALM}/roles`, role);
    if (!res.ok) {
      const text = await res.text();
      throw new Error(`Falha ao criar role '${role.name}' (${res.status}): ${text}`);
    }
    log(`  Role '${role.name}' criada.`);
  }
}

async function ensureClientScopes(token: string): Promise<void> {
  log("Configurando client scopes...");

  const listRes = await adminFetch(token, "GET", `/admin/realms/${REALM}/client-scopes`);
  if (!listRes.ok) {
    throw new Error(`Falha ao listar client scopes (${listRes.status})`);
  }
  const existing = (await listRes.json()) as Array<{ id: string; name: string }>;
  const existingMap = new Map(existing.map((s) => [s.name, s.id]));

  for (const scope of CLIENT_SCOPES) {
    if (existingMap.has(scope.name)) {
      log(`  Client scope '${scope.name}' já existe. Pulando.`);
      continue;
    }
    const res = await adminFetch(token, "POST", `/admin/realms/${REALM}/client-scopes`, scope);
    if (!res.ok) {
      const text = await res.text();
      throw new Error(`Falha ao criar client scope '${scope.name}' (${res.status}): ${text}`);
    }
    log(`  Client scope '${scope.name}' criado.`);
  }
}

async function ensureClient(token: string): Promise<void> {
  log(`Configurando cliente '${CONNECT_CLIENT.clientId}'...`);

  const listRes = await adminFetch(token, "GET", `/admin/realms/${REALM}/clients?clientId=${CONNECT_CLIENT.clientId}`);
  if (!listRes.ok) {
    throw new Error(`Falha ao listar clientes (${listRes.status})`);
  }
  const existing = (await listRes.json()) as Array<{ id: string; clientId: string }>;

  if (existing.length > 0) {
    log(`  Cliente '${CONNECT_CLIENT.clientId}' já existe. Pulando.`);
    return;
  }

  const res = await adminFetch(token, "POST", `/admin/realms/${REALM}/clients`, CONNECT_CLIENT);
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Falha ao criar cliente (${res.status}): ${text}`);
  }
  log(`  Cliente '${CONNECT_CLIENT.clientId}' criado.`);
}

async function getRealmRoleByName(token: string, roleName: string): Promise<RealmRole> {
  const res = await adminFetch(token, "GET", `/admin/realms/${REALM}/roles/${encodeURIComponent(roleName)}`);
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Falha ao buscar role '${roleName}' (${res.status}): ${text}`);
  }
  return res.json() as Promise<RealmRole>;
}

async function ensureUsers(token: string): Promise<void> {
  log("Configurando usuários de teste...");

  for (const user of TEST_USERS) {
    // Verifica se o usuário já existe
    const searchRes = await adminFetch(
      token,
      "GET",
      `/admin/realms/${REALM}/users?username=${encodeURIComponent(user.username)}&exact=true`,
    );
    if (!searchRes.ok) {
      throw new Error(`Falha ao buscar usuário '${user.username}' (${searchRes.status})`);
    }
    const found = (await searchRes.json()) as Array<{ id: string; username: string }>;

    let userId: string;

    if (found.length > 0) {
      userId = found[0].id;
      log(`  Usuário '${user.username}' já existe (${userId}).`);
    } else {
      // Cria o usuário
      const createRes = await adminFetch(token, "POST", `/admin/realms/${REALM}/users`, {
        username: user.username,
        email: user.email,
        firstName: user.firstName,
        lastName: user.lastName,
        enabled: true,
        emailVerified: true,
        credentials: [{ type: "password", value: user.password, temporary: false }],
        attributes: { organization: [ORG_ATTRIBUTE] },
      });

      if (!createRes.ok) {
        const text = await createRes.text();
        throw new Error(`Falha ao criar usuário '${user.username}' (${createRes.status}): ${text}`);
      }

      // Extrai o ID do header Location
      const location = createRes.headers.get("Location") ?? "";
      const parts = location.replace(/\/$/, "").split("/");
      userId = parts[parts.length - 1];
      log(`  Usuário '${user.username}' criado (${userId}).`);
    }

    // Atribui roles
    const roleObjects: RealmRole[] = await Promise.all(
      user.roles.map((r) => getRealmRoleByName(token, r)),
    );

    const rolesRes = await adminFetch(
      token,
      "POST",
      `/admin/realms/${REALM}/users/${userId}/role-mappings/realm`,
      roleObjects,
    );
    if (!rolesRes.ok) {
      const text = await rolesRes.text();
      // 409 pode ocorrer se a role já está atribuída — não é erro fatal
      if (rolesRes.status !== 409) {
        throw new Error(`Falha ao atribuir roles ao usuário '${user.username}' (${rolesRes.status}): ${text}`);
      }
    }
    log(`  Roles [${user.roles.join(", ")}] atribuídas ao usuário '${user.username}'.`);
  }
}

// ---------------------------------------------------------------------------
// Utilitários
// ---------------------------------------------------------------------------

function log(msg: string): void {
  const ts = new Date().toISOString();
  console.log(`[${ts}] ${msg}`);
}

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

async function main(): Promise<void> {
  log("=== Keycloak Setup — Realm Luxus ===");
  log(`URL: ${KEYCLOAK_URL}`);
  log(`Admin: ${ADMIN_USER}`);
  log(`Realm: ${REALM}`);
  log("");

  const token = await getAdminToken();

  await ensureRealm(token);
  await ensureRealmRoles(token);
  await ensureClientScopes(token);
  await ensureClient(token);
  await ensureUsers(token);

  log("");
  log("=== Setup concluído com sucesso! ===");
}

main().catch((err) => {
  console.error("[ERRO]", err instanceof Error ? err.message : err);
  process.exit(1);
});
