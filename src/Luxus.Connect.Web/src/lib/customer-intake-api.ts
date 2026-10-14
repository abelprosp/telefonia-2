import { useQuery } from '@tanstack/react-query';

import { client } from '@/lib/client';

export type ImportedSale = {
  id: string;
  protocol: string;
  taskProtocol: string;
  soldAt: string;
  concludedAt: string | null;
  value: number;
  commissionValue: number;
  operatorName: string;
  planName: string;
  lineNumber: string;
  simType: string;
  isVirginChip: boolean;
  iccid: string;
  deviceImei: string;
  deviceEid: string;
  devicePlatform: string;
  isPortability: boolean;
  portabilityNumber: string;
  donorOperator: string;
  isUpgrade: boolean;
  upgradeOfProtocol: string;
  contractFormat: string;
  partnerName: string;
  branchName: string;
  notes: string;
};

export type CustomerIntake = {
  id: string;
  source_system: string;
  customer_id: string;
  sale_id: string;
  imported_at: string;
  sale: ImportedSale;
  client: {
    id?: string;
    name: string;
    document: string;
    documentType: string;
    legalName?: string;
    rg: string;
    email: string;
    phone: string;
    address: string;
    addressNumber: string;
    complement: string;
    neighborhood: string;
    city: string;
    state: string;
    zipCode: string;
  };
  attachments: {
    external_id: string;
    attachment_id: string;
    type: string;
    name: string;
    mime_type: string;
    size_bytes: number;
  }[];
};

export function useCustomerIntakes(customerId: string) {
  return useQuery({
    queryKey: ['customer-intakes', customerId],
    queryFn: async () =>
      (
        await client<CustomerIntake[]>({
          url: `/v1/customers/${customerId}/imported-sales`,
          method: 'GET'
        })
      ).data,
    enabled: Boolean(customerId)
  });
}
