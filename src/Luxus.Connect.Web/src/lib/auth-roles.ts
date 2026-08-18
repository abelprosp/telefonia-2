import { useMemo } from 'react';

import { useAuth } from 'react-oidc-context';

import { decode } from '@/lib/jwt';
import { type TokenPayload } from '@/types/auth';

export type UserProfile =
  | 'master'
  | 'employee'
  | 'operator'
  | 'financial'
  | 'sales'
  | 'viewer'
  | 'partner'
  | 'user';

export type AuthRoleState = {
  roles: string[];
  isMaster: boolean;
  isEmployee: boolean;
  isOperator: boolean;
  isFinancial: boolean;
  isSales: boolean;
  isViewer: boolean;
  isPartner: boolean;
  isCustomerPortal: boolean;
  isInternalStaff: boolean;
  isPartnerOnly: boolean;
  canAccessOperations: boolean;
  canAccessFinance: boolean;
  canManageUsers: boolean;
  profile: UserProfile;
};

function rolesFromAccessToken(token: string | undefined): string[] {
  if (!token) {
    return [];
  }

  const payload = decode<TokenPayload & { roles?: string[] }>(token);
  if (!payload) {
    return [];
  }

  const fromClaim = payload.roles ?? [];
  const fromRealm = payload.realm_access?.roles ?? [];
  const fromClient = Object.values(payload.resource_access ?? {}).flatMap(
    (entry) => entry.roles ?? []
  );

  return [...new Set([...fromClaim, ...fromRealm, ...fromClient])];
}

export function resolveProfile(roles: string[]): UserProfile {
  if (roles.includes('master') || roles.includes('admin')) return 'master';
  if (roles.includes('financial')) return 'financial';
  if (roles.includes('sales')) return 'sales';
  if (roles.includes('operator')) return 'operator';
  if (roles.includes('employee')) return 'employee';
  if (roles.includes('viewer')) return 'viewer';
  if (roles.includes('partner')) return 'partner';
  return 'user';
}

export function profileLabel(profile: UserProfile): string {
  const map: Record<UserProfile, string> = {
    master: 'Master',
    employee: 'Funcionário',
    operator: 'Operador',
    financial: 'Financeiro',
    sales: 'Comercial',
    viewer: 'Consulta',
    partner: 'Parceiro',
    user: 'Usuário'
  };
  return map[profile];
}

export function useAuthRoles(): AuthRoleState {
  const { user } = useAuth();

  return useMemo(() => {
    const roles = rolesFromAccessToken(user?.access_token);
    const isMaster = roles.includes('master') || roles.includes('admin');
    const isEmployee = roles.includes('employee');
    const isOperator = roles.includes('operator');
    const isFinancial = roles.includes('financial');
    const isSales = roles.includes('sales');
    const isViewer = roles.includes('viewer');
    const isPartner = roles.includes('partner');
    const isInternalStaff = isMaster || isEmployee || isOperator || isFinancial || isSales || isViewer;
    const isCustomerPortal = !isInternalStaff && !isPartner;
    const profile = resolveProfile(roles);

    return {
      roles,
      isMaster,
      isEmployee,
      isOperator,
      isFinancial,
      isSales,
      isViewer,
      isPartner,
      isCustomerPortal,
      isInternalStaff,
      isPartnerOnly: isPartner && !isInternalStaff,
      canAccessOperations: isMaster || isEmployee || isOperator || isSales,
      canAccessFinance: isMaster || isFinancial,
      canManageUsers: isMaster,
      profile
    };
  }, [user?.access_token]);
}

export function roleLabel(roles: Pick<AuthRoleState, 'profile'>) {
  return profileLabel(roles.profile);
}

export const PROFILE_OPTIONS: { value: UserProfile; label: string; description: string }[] = [
  {
    value: 'master',
    label: 'Master',
    description: 'Acesso total ao sistema e gestão de usuários'
  },
  {
    value: 'employee',
    label: 'Funcionário',
    description: 'Operação: clientes, linhas, faturamento e vendas (sem financeiro)'
  },
  {
    value: 'operator',
    label: 'Operador',
    description: 'Operação de linhas, faturamento e divergências'
  },
  {
    value: 'financial',
    label: 'Financeiro',
    description: 'Controle financeiro completo: contas, comissões e relatórios'
  },
  {
    value: 'sales',
    label: 'Comercial',
    description: 'Vendas e contratos, sem acesso financeiro'
  },
  {
    value: 'viewer',
    label: 'Consulta',
    description: 'Leitura operacional sem mutações privilegiadas'
  },
  {
    value: 'partner',
    label: 'Parceiro',
    description: 'Portal do parceiro: clientes, linhas e vendas da carteira'
  },
  {
    value: 'user',
    label: 'Usuário',
    description: 'Portal do cliente (CPF/CNPJ)'
  }
];
