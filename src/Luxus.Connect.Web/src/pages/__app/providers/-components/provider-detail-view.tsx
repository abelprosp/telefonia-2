import { Link } from '@tanstack/react-router';
import { ChevronDown, ChevronLeft, ChevronRight, Plus, Users } from 'lucide-react';

import type { GetProviderPlanResponse, GetProviderResponse } from '@/api';
import { Button } from '@/components/ui/button';
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Separator } from '@/components/ui/separator';

import { ProviderPlanServices } from './provider-plan-services';

type ProvidersListSearch = {
  page: number;
  pageSize: number;
};

function DetailSection({
  title,
  description,
  children
}: {
  title: string;
  description: string;
  children: React.ReactNode;
}) {
  return (
    <div className="grid grid-cols-1 gap-8 md:grid-cols-3">
      <div>
        <h2 className="text-foreground font-semibold">{title}</h2>
        <p className="text-muted-foreground mt-1 text-sm leading-6">
          {description}
        </p>
      </div>
      <div className="sm:max-w-3xl md:col-span-2">{children}</div>
    </div>
  );
}

function ReadOnlyField({ value }: { value: string }) {
  return (
    <Input
      readOnly
      value={value}
      className="bg-muted/50 pointer-events-none border-transparent shadow-none"
    />
  );
}

type ProviderDetailViewProps = {
  provider: GetProviderResponse;
  providerId: string;
  listSearch: ProvidersListSearch;
  openPlanId: string | null;
  onTogglePlan: (planId: string | null) => void;
  onAddPlan?: () => void;
};

export function ProviderDetailView({
  provider,
  providerId,
  listSearch,
  openPlanId,
  onTogglePlan,
  onAddPlan
}: ProviderDetailViewProps) {
  const backLink = {
    to: '/providers' as const,
    search: listSearch
  };

  const planItems: GetProviderPlanResponse[] = provider.plans ?? [];

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="flex items-center gap-4">
          <Button
            nativeButton={false}
            variant="outline"
            size="icon"
            render={<Link {...backLink} />}
          >
            <ChevronLeft className="size-4" />
            <span className="sr-only">Voltar</span>
          </Button>
          <div className="flex flex-col">
            <h3 className="text-foreground text-lg font-semibold">Operadora</h3>
            <p className="text-muted-foreground mt-1 text-sm leading-6">
              {provider.name}
            </p>
          </div>
        </div>
      </div>

      <DetailSection
        title="Dados da operadora"
        description="Identificação e status cadastrados no sistema."
      >
        <FieldGroup className="gap-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field>
              <FieldLabel>Nome</FieldLabel>
              <ReadOnlyField value={provider.name} />
            </Field>
            <Field>
              <FieldLabel>Slug</FieldLabel>
              <ReadOnlyField value={provider.slug} />
            </Field>
            <Field>
              <FieldLabel>ID</FieldLabel>
              <ReadOnlyField value={provider.id} />
            </Field>
            <Field>
              <FieldLabel>Ativa</FieldLabel>
              <ReadOnlyField value={provider.active ? 'Sim' : 'Não'} />
            </Field>
          </div>
        </FieldGroup>
      </DetailSection>

      <Separator />

      <DetailSection
        title="Clientes"
        description="Acesse a listagem de clientes filtrada por esta operadora."
      >
        <div className="flex flex-wrap gap-4">
          <Link
            to="/customers"
            search={{
              page: 1,
              pageSize: 10,
              providerId
            }}
            className="border-border bg-card hover:bg-accent/40 focus-visible:ring-ring flex min-w-40 flex-col gap-1 rounded-lg border px-4 py-3 text-sm transition-colors focus-visible:ring-2 focus-visible:outline-none"
          >
            <span className="text-muted-foreground flex items-center gap-2">
              <Users className="size-4" />
              Ver clientes
            </span>
            <span className="text-muted-foreground text-xs">
              Listagem filtrada por operadora
            </span>
          </Link>
        </div>
      </DetailSection>

      <Separator />

      <DetailSection
        title="Planos e serviços"
        description="Cadastre planos para vincular às linhas do estoque. Expanda um plano para ver os serviços e o valor mensal."
      >
        <div className="mb-4 flex flex-wrap justify-end">
          <Button type="button" size="sm" onClick={onAddPlan}>
            <Plus className="mr-2 size-4" />
            Novo plano
          </Button>
        </div>
        {planItems.length === 0 ? (
          <p className="text-muted-foreground text-sm">
            Nenhum plano cadastrado. Clique em &quot;Novo plano&quot; para cadastrar e usar no estoque de
            linhas.
          </p>
        ) : (
          <div className="flex flex-col gap-2">
            {planItems.map((plan) => {
              const isOpen = openPlanId === plan.id;
              return (
                <div key={plan.id} className="border-border rounded-lg border">
                  <button
                    type="button"
                    className="hover:bg-accent/50 flex w-full items-center justify-between gap-3 px-4 py-3 text-left text-sm"
                    onClick={() => {
                      onTogglePlan(isOpen ? null : plan.id);
                    }}
                  >
                    <div>
                      <div className="font-medium">{plan.name}</div>
                      <div className="text-muted-foreground text-xs">
                        Código: {plan.code}
                      </div>
                    </div>
                    {isOpen ? (
                      <ChevronDown className="text-muted-foreground size-4 shrink-0" />
                    ) : (
                      <ChevronRight className="text-muted-foreground size-4 shrink-0" />
                    )}
                  </button>
                  {isOpen && (
                    <div className="border-border border-t">
                      <ProviderPlanServices services={plan.services ?? []} />
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </DetailSection>

      <div className="flex flex-wrap items-center justify-end gap-4">
        <Button
          nativeButton={false}
          type="button"
          variant="outline"
          className="whitespace-nowrap"
          render={<Link {...backLink} />}
        >
          Voltar
        </Button>
      </div>
    </div>
  );
}
