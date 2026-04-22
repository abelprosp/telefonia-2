# Luxus Connect — Web

Frontend em **Vite + React + TypeScript**, baseado no template **ritter-vite-starter** (TanStack Router, React Query, Tailwind 4, shadcn/ui).

## Template legado (Next.js)

No monorepo, em **`../old`**, está a **versão anterior do frontend em Next.js**. Esse diretório não é a aplicação ativa; serve como **template de referência** para consultar implementações, fluxos e componentes ao portar ou recriar telas no projeto atual.

## Scripts

- `pnpm dev` — desenvolvimento (http://localhost:5173)
- `pnpm build` — `tsc -b` + build de produção
- `pnpm preview` — pré-visualizar o build
- `pnpm lint` — ESLint

## Variáveis de ambiente

Copie `.env.example` para `.env`. Obrigatórios para o cliente: `VITE_API_URL`, `VITE_AUTH_URL` (ver `src/env.ts`).

Alinhar `VITE_API_URL` com a API **Luxus.Connect.Api** (ex.: `http://localhost:5193`) e CORS para a origem do Vite.

## shadcn/ui

Configuração em `components.json`. Adicionar componentes: `npx shadcn@latest add <nome>`.
