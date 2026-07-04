const roles = {
    realm: [
        { name: "admin", description: "Administrador Luxus Connect" },
        { name: "user", description: "Utilizador padrão" },
        { name: "partner", description: "Parceiro comercial com acesso restrito" },
        { name: "master", description: "Acesso total ao sistema e gestão de usuários" },
        { name: "employee", description: "Funcionário — operação básica sem módulo financeiro" },
        { name: "financial", description: "Financeiro — controle financeiro completo" },
    ],
};

const users = [
    {
        username: "dev",
        enabled: true,
        emailVerified: true,
        firstName: "Dev",
        lastName: "Luxus",
        email: "dev@luxus.local",
        credentials: [{ type: "password", value: "dev", temporary: false }],
        realmRoles: ["master", "admin", "user"],
    },
    {
        username: "parceiro",
        enabled: true,
        emailVerified: true,
        firstName: "Parceiro",
        lastName: "Luxus",
        email: "parceiro@luxus.local",
        credentials: [{ type: "password", value: "parceiro", temporary: false }],
        realmRoles: ["partner", "user"],
    },
    {
        username: "funcionario",
        enabled: true,
        emailVerified: true,
        firstName: "Funcionario",
        lastName: "Luxus",
        email: "funcionario@luxus.local",
        credentials: [{ type: "password", value: "funcionario", temporary: false }],
        realmRoles: ["employee", "user"],
    },
    {
        username: "financeiro",
        enabled: true,
        emailVerified: true,
        firstName: "Financeiro",
        lastName: "Luxus",
        email: "financeiro@luxus.local",
        credentials: [{ type: "password", value: "financeiro", temporary: false }],
        realmRoles: ["financial", "user"],
    },
];


const clientScopes = [
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

const clients = [
    {
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
            `${process.env.CONNECT_WEB_URL}/*` || "https://connect-web-production-e247.up.railway.app/*",
        ],
        webOrigins: [
            "http://localhost:5173",
            "http://localhost:3000",
            "http://localhost:8002",
            "http://127.0.0.1:5173",
            process.env.CONNECT_WEB_URL || "https://connect-web-production-e247.up.railway.app",
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
    },
];

const realmJson = {
    realm: "luxus",
    enabled: true,
    registrationAllowed: false,
    loginWithEmailAllowed: true,
    duplicateEmailsAllowed: false,
    resetPasswordAllowed: true,
    editUsernameAllowed: false,
    bruteForceProtected: false,
    roles,
    defaultRoles: ["user"],
    users,
    clientScopes,
    clients,
};

async function getAccessToken(keycloakUrl, adminUser, adminPassword) {
    console.log(`Conectando ao Keycloak em ${keycloakUrl}...`);

    // Obter token de admin
    const tokenResponse = await fetch(`${keycloakUrl}/realms/master/protocol/openid-connect/token`, {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: new URLSearchParams({
            grant_type: "password",
            client_id: "admin-cli",
            username: adminUser,
            password: adminPassword,
        }),
    });

    if (!tokenResponse.ok) {
        throw new Error(`Falha ao obter token: ${tokenResponse.status} ${tokenResponse.statusText}`);
    }

    const { access_token } = await tokenResponse.json();
    console.log("Token obtido");

    return access_token;
}

async function importRealm(keycloakUrl, accessToken) {
    try {
        // Verificar se realm já existe
        const realmCheckResponse = await fetch(`${keycloakUrl}/admin/realms/luxus`, {
            headers: { Authorization: `Bearer ${accessToken}` },
        });

        if (realmCheckResponse.ok) {
            console.warn("Realm 'luxus' já existe. Atualizando...");

            const updateResponse = await fetch(`${keycloakUrl}/admin/realms/luxus`, {
                method: "PUT",
                headers: {
                    "Content-Type": "application/json",
                    Authorization: `Bearer ${accessToken}`,
                },
                body: JSON.stringify(realmJson),
            });

            if (!updateResponse.ok) {
                throw new Error(`Falha ao atualizar realm: ${updateResponse.status}`);
            }

            console.log("Realm atualizado");
        } else {
            // Criar novo realm
            const createResponse = await fetch(`${keycloakUrl}/admin/realms`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    Authorization: `Bearer ${accessToken}`,
                },
                body: JSON.stringify(realmJson),
            });

            if (!createResponse.ok) {
                throw new Error(`Falha ao criar realm: ${createResponse.status}`);
            }

            console.log("Realm 'luxus' criado com sucesso");
        }

        console.log("\nUsuários de teste:");
        console.log("  dev / dev (master)");
        console.log("  parceiro / parceiro (partner)");
        console.log("  funcionario / funcionario (employee)");
        console.log("  financeiro / financeiro (financial)");
    } catch (error) {
        console.error("Erro:", error instanceof Error ? error.message : error);
        process.exit(1);
    }
}

(async function main() {
    const keycloakUrl = process.env.KEYCLOAK_URL;
    const adminUser = process.env.KC_BOOTSTRAP_ADMIN_USERNAME;
    const adminPassword = process.env.KC_BOOTSTRAP_ADMIN_PASSWORD;

    const notDefined = (v) => [undefined, null, ""].includes(v);

    if (notDefined(keycloakUrl) || notDefined(adminUser) || notDefined(adminPassword)) {
        console.warn('Credenciais para acesso ao Keycloak não informadas! Favor verificar variáveis de ambiente.');
        process.exit(1);
    }

    const accessToken = await getAccessToken(keycloakUrl, adminUser, adminPassword);

    await importRealm(keycloakUrl, accessToken);
})();
