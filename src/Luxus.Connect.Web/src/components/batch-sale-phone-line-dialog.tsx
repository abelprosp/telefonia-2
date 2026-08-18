import { useMemo, useState } from 'react';

import { CheckSquare, Filter, Layers, Plus, Search, Square } from 'lucide-react';

import { useGetV1PhoneLines, type ListPhoneLineResponse } from '@/api';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { formatPhoneNumber } from '@/lib/format';
import { formatMoney } from '@/lib/sales-api';

const ALL_PLANS = '__all__';

type BatchSalePhoneLineDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  excludeLineIds?: string[];
  onAddLines: (
    lines: Array<{
      line: ListPhoneLineResponse;
      price: number;
    }>
  ) => void;
};

export function BatchSalePhoneLineDialog({
  open,
  onOpenChange,
  excludeLineIds = [],
  onAddLines
}: BatchSalePhoneLineDialogProps) {
  const [planFilter, setPlanFilter] = useState(ALL_PLANS);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedLineIds, setSelectedLineIds] = useState<Set<string>>(new Set());
  const [overridePrice, setOverridePrice] = useState<string>('');
  const [useCustomPrice, setUseCustomPrice] = useState(false);

  const stockQuery = useGetV1PhoneLines({
    page_index: 0,
    page_size: 500,
    status: 'in_stock'
  });

  const availableLines = useMemo(() => {
    const excluded = new Set(excludeLineIds);
    return (stockQuery.data?.items ?? []).filter((line) => !excluded.has(line.id));
  }, [excludeLineIds, stockQuery.data?.items]);

  const planOptions = useMemo(() => {
    const map = new Map<string, string>();
    for (const line of availableLines) {
      if (!map.has(line.provider_plan_id)) {
        map.set(line.provider_plan_id, line.provider_plan_name);
      }
    }
    return Array.from(map.entries())
      .map(([id, name]) => ({ id, name }))
      .sort((a, b) => a.name.localeCompare(b.name, 'pt-BR'));
  }, [availableLines]);

  const filteredLines = useMemo(() => {
    return availableLines.filter((line) => {
      if (planFilter !== ALL_PLANS && line.provider_plan_id !== planFilter) {
        return false;
      }
      if (searchTerm.trim()) {
        const cleanSearch = searchTerm.replace(/\D/g, '');
        const cleanNumber = (line.number ?? '').replace(/\D/g, '');
        if (cleanSearch && !cleanNumber.includes(cleanSearch)) {
          return false;
        }
        if (
          !cleanSearch &&
          !line.number.toLowerCase().includes(searchTerm.toLowerCase()) &&
          !line.provider_plan_name.toLowerCase().includes(searchTerm.toLowerCase())
        ) {
          return false;
        }
      }
      return true;
    });
  }, [availableLines, planFilter, searchTerm]);

  const toggleSelectLine = (id: string) => {
    setSelectedLineIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const selectAllFiltered = () => {
    setSelectedLineIds((prev) => {
      const next = new Set(prev);
      for (const l of filteredLines) {
        next.add(l.id);
      }
      return next;
    });
  };

  const clearSelection = () => {
    setSelectedLineIds(new Set());
  };

  const handleConfirmAdd = () => {
    const selectedLines = availableLines.filter((l) => selectedLineIds.has(l.id));
    const customNum = parseFloat(overridePrice.replace(',', '.'));

    const payload = selectedLines.map((line) => {
      let finalPrice = 0;
      if (useCustomPrice && !isNaN(customNum) && customNum >= 0) {
        finalPrice = customNum;
      } else if (line.base_cost != null && line.base_cost > 0) {
        finalPrice = Number(line.base_cost);
      }
      return { line, price: finalPrice };
    });

    onAddLines(payload);
    setSelectedLineIds(new Set());
    onOpenChange(false);
  };

  const allFilteredSelected =
    filteredLines.length > 0 &&
    filteredLines.every((l) => selectedLineIds.has(l.id));

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="flex w-full flex-col sm:max-w-2xl overflow-y-auto">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2">
            <Layers className="text-primary size-5" />
            Adicionar Linhas em Lote à Venda
          </SheetTitle>
          <SheetDescription>
            Selecione múltiplas linhas telefônicas disponíveis em estoque para adicionar de uma só vez aos itens da venda.
          </SheetDescription>
        </SheetHeader>

        <div className="flex flex-col gap-4 py-3">
          {/* Filtros e Busca */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label className="text-xs flex items-center gap-1">
                <Filter className="size-3 text-muted-foreground" /> Filtrar por Plano
              </Label>
              <Select
                value={planFilter}
                onValueChange={(v) => setPlanFilter(v ?? ALL_PLANS)}
              >
                <SelectTrigger className="h-9 text-xs">
                  <SelectValue placeholder="Todos os planos" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={ALL_PLANS}>Todos os planos ({availableLines.length})</SelectItem>
                  {planOptions.map((p) => (
                    <SelectItem key={p.id} value={p.id}>
                      {p.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1.5">
              <Label className="text-xs flex items-center gap-1">
                <Search className="size-3 text-muted-foreground" /> Buscar Número
              </Label>
              <Input
                placeholder="Ex: 11999998888 ou DDD..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="h-9 text-xs"
              />
            </div>
          </div>

          {/* Configuração de Preço em Lote */}
          <div className="rounded-2xl border bg-muted/30 p-3.5 space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold text-foreground">
                Regra de Preço Unitário para o Lote:
              </span>
              <label className="flex items-center gap-1.5 text-xs font-medium cursor-pointer text-foreground">
                <input
                  type="checkbox"
                  checked={useCustomPrice}
                  onChange={(e) => setUseCustomPrice(e.target.checked)}
                  className="rounded border-gray-300"
                />
                Definir valor fixo para todas
              </label>
            </div>

            {useCustomPrice ? (
              <div className="flex items-center gap-2 pt-1">
                <Label className="text-xs whitespace-nowrap">Valor por Linha (R$):</Label>
                <Input
                  type="number"
                  step="0.01"
                  min="0"
                  placeholder="Ex: 49.90"
                  value={overridePrice}
                  onChange={(e) => setOverridePrice(e.target.value)}
                  className="h-8 text-xs max-w-[140px]"
                />
                <span className="text-muted-foreground text-[11px]">
                  (Substitui o valor individual das linhas selecionadas)
                </span>
              </div>
            ) : (
              <p className="text-muted-foreground text-[11px]">
                O sistema usará automaticamente o valor cadastrado de cada linha em estoque (ou R$ 0,00 se não houver).
              </p>
            )}
          </div>

          {/* Cabeçalho da Tabela com Selecionar Tudo */}
          <div className="flex items-center justify-between border-b pb-2">
            <div className="flex items-center gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={allFilteredSelected ? clearSelection : selectAllFiltered}
                className="h-7 gap-1 text-xs"
              >
                {allFilteredSelected ? (
                  <>
                    <Square className="size-3" /> Desmarcar Visíveis
                  </>
                ) : (
                  <>
                    <CheckSquare className="size-3" /> Selecionar Todas Visíveis ({filteredLines.length})
                  </>
                )}
              </Button>
              {selectedLineIds.size > 0 ? (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={clearSelection}
                  className="h-7 text-xs text-muted-foreground"
                >
                  Limpar ({selectedLineIds.size})
                </Button>
              ) : null}
            </div>

            <Badge variant="outline" className="text-xs">
              {selectedLineIds.size} de {availableLines.length} selecionadas
            </Badge>
          </div>

          {/* Lista de Linhas */}
          {stockQuery.isPending ? (
            <p className="text-muted-foreground py-6 text-center text-xs">
              Carregando estoque de linhas...
            </p>
          ) : filteredLines.length === 0 ? (
            <div className="rounded-xl border border-dashed p-6 text-center text-xs text-muted-foreground">
              Nenhuma linha encontrada para os filtros selecionados.
            </div>
          ) : (
            <div className="max-h-[340px] overflow-y-auto rounded-xl border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-10 text-center"></TableHead>
                    <TableHead>Número</TableHead>
                    <TableHead>Plano</TableHead>
                    <TableHead className="text-right">Valor Padrão</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredLines.map((line) => {
                    const isSelected = selectedLineIds.has(line.id);
                    const formatted = formatPhoneNumber(line.number) ?? line.number;
                    const price = line.base_cost != null && line.base_cost > 0 ? Number(line.base_cost) : 0;

                    return (
                      <TableRow
                        key={line.id}
                        className={`cursor-pointer transition-colors ${
                          isSelected ? 'bg-primary/5 font-medium' : 'hover:bg-muted/40'
                        }`}
                        onClick={() => toggleSelectLine(line.id)}
                      >
                        <TableCell className="text-center" onClick={(e) => e.stopPropagation()}>
                          <input
                            type="checkbox"
                            checked={isSelected}
                            onChange={() => toggleSelectLine(line.id)}
                            className="size-4 rounded border-gray-300"
                          />
                        </TableCell>
                        <TableCell className="font-mono text-xs text-foreground font-semibold">
                          {formatted}
                        </TableCell>
                        <TableCell className="text-xs text-muted-foreground">
                          {line.provider_plan_name}
                        </TableCell>
                        <TableCell className="text-right font-mono text-xs text-foreground">
                          {price > 0 ? formatMoney(price) : '—'}
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
          )}
        </div>

        <SheetFooter className="mt-auto border-t pt-4">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancelar
          </Button>
          <Button
            onClick={handleConfirmAdd}
            disabled={selectedLineIds.size === 0}
            className="gap-1.5"
          >
            <Plus className="size-4" />
            Adicionar {selectedLineIds.size} {selectedLineIds.size === 1 ? 'Linha' : 'Linhas'} à Venda
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}
