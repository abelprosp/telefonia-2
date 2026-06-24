import { useMemo, useState } from 'react';

import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { type DeviceStockItem, useDeviceStockList } from '@/lib/device-stock-api';
import { formatMoney } from '@/lib/sales-api';

const ALL_BRANDS = '__all__';

type SaleDevicePickerProps = {
  value?: string;
  excludeSkus?: string[];
  onChange: (device: DeviceStockItem | null) => void;
};

function deviceLabel(device: DeviceStockItem) {
  const parts = [`${device.brand} ${device.model}`, device.sku];
  if (device.storage_capacity) parts.push(device.storage_capacity);
  if (device.color) parts.push(device.color);
  if (device.sale_price != null) parts.push(formatMoney(device.sale_price));
  return parts.join(' · ');
}

export function SaleDevicePicker({ value, excludeSkus = [], onChange }: SaleDevicePickerProps) {
  const [brandFilter, setBrandFilter] = useState(ALL_BRANDS);

  const stockQuery = useDeviceStockList({
    page_index: 0,
    page_size: 200,
    status: 'in_stock'
  });

  const availableDevices = useMemo(() => {
    const excluded = new Set(excludeSkus);
    return (stockQuery.data?.items ?? []).filter((device) => !excluded.has(device.sku));
  }, [excludeSkus, stockQuery.data?.items]);

  const brandOptions = useMemo(() => {
    const brands = new Set(availableDevices.map((device) => device.brand));
    return Array.from(brands).sort((a, b) => a.localeCompare(b, 'pt-BR'));
  }, [availableDevices]);

  const filteredDevices = useMemo(() => {
    if (brandFilter === ALL_BRANDS) return availableDevices;
    return availableDevices.filter((device) => device.brand === brandFilter);
  }, [availableDevices, brandFilter]);

  const selectedDevice = useMemo(
    () => availableDevices.find((device) => device.sku === value) ?? null,
    [availableDevices, value]
  );

  const handleBrandChange = (next: string) => {
    setBrandFilter(next);
    if (value && next !== ALL_BRANDS) {
      const stillVisible = availableDevices.some(
        (device) => device.sku === value && device.brand === next
      );
      if (!stillVisible) {
        onChange(null);
      }
    }
  };

  if (stockQuery.isPending) {
    return <p className="text-muted-foreground text-sm">Carregando aparelhos em estoque…</p>;
  }

  if (availableDevices.length === 0) {
    return (
      <p className="text-muted-foreground text-sm">
        Nenhum aparelho em estoque disponível. Cadastre aparelhos em Cadastros → Estoque de aparelhos.
      </p>
    );
  }

  return (
    <div className="space-y-3 rounded-lg border border-dashed p-3">
      <p className="text-muted-foreground text-xs">
        Selecione o aparelho disponível no estoque.
      </p>

      {brandOptions.length > 1 && (
        <div className="space-y-2">
          <Label>Marca</Label>
          <Select value={brandFilter} onValueChange={(v) => handleBrandChange(v ?? ALL_BRANDS)}>
            <SelectTrigger>
              <SelectValue placeholder="Todas as marcas" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL_BRANDS}>Todas as marcas</SelectItem>
              {brandOptions.map((brand) => (
                <SelectItem key={brand} value={brand}>
                  {brand}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}

      <div className="space-y-2">
        <Label>Aparelho em estoque</Label>
        {filteredDevices.length === 0 ? (
          <p className="text-muted-foreground text-sm">Nenhum aparelho nesta marca.</p>
        ) : (
          <Select
            value={value ?? ''}
            onValueChange={(sku) => {
              const device = filteredDevices.find((d) => d.sku === sku) ?? null;
              onChange(device);
            }}
          >
            <SelectTrigger>
              <SelectValue placeholder="Selecione o aparelho" />
            </SelectTrigger>
            <SelectContent>
              {filteredDevices.map((device) => (
                <SelectItem key={device.id} value={device.sku}>
                  {deviceLabel(device)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </div>

      {selectedDevice?.imei && (
        <p className="text-muted-foreground text-xs">IMEI: {selectedDevice.imei}</p>
      )}
    </div>
  );
}

export function buildDeviceSaleDescription(device: DeviceStockItem) {
  const model = `${device.brand} ${device.model}`;
  const extras = [device.storage_capacity, device.color].filter(Boolean).join(' ');
  return extras ? `Aparelho ${model} ${extras}` : `Aparelho ${model}`;
}
