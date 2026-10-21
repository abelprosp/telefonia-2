# Luxus Connect — Web

Frontend em **Vite + React + TypeScript**, baseado no template **ritter-vite-starter** (TanStack Router, React Query, Tailwind 4, shadcn/ui).

## Template legado (Next.js)

No monorepo, em **`../old`**, está a **versão anterior do frontend em Next.js**. Esse diretório não é a aplicação ativa; serve como **template de referência** para consultar implementações, fluxos e componentes ao portar ou recriar telas no projeto atual.

## Scripts

- `npm run dev` — desenvolvimento (http://localhost:5173)
- `npm run build` — `tsc -b` + build de produção
- `npm run preview` — pré-visualizar o build
- `npm run lint` — ESLint

## Variáveis de ambiente

Copie `.env.example` para `.env`. Obrigatórios para o cliente: `VITE_API_URL`, `VITE_STORAGE_BUCKET_NAME` (ver `src/env.ts`).

Alinhar `VITE_API_URL` com a API Go (ex.: `http://localhost:8002`) e CORS
para a origem do Vite.

## Login e autenticação em dois fatores

O login usa a API do Telefonia e uma sessão em cookie HttpOnly. Não há provedor externo nem tokens persistidos no navegador. Em produção, publique frontend e API com HTTPS na mesma origem.

O usuário entra com senha e confirma o acesso com uma passkey (digital, Face ID, Windows Hello ou chave de segurança). No primeiro acesso que exige MFA, o próprio Telefonia cadastra a passkey e apresenta os códigos de recuperação. Guarde estes códigos em local seguro: cada um funciona uma única vez. A tela de login permite alternar entre a passkey e um código de recuperação.

## shadcn/ui

Configuração em `components.json`. Os componentes Base UI e suas variantes CSS são mantidos localmente; o CLI shadcn não faz parte das dependências da aplicação.

Veja [as correções e os passos de produção](../../docs/SECURITY_HARDENING.md). Para validar a sanitização dos contratos, execute `npm run test:security`.
