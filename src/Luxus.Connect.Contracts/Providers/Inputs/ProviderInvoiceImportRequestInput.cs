using Luxus.Connect.Contracts.Providers.Commands;

namespace Luxus.Connect.Contracts.Providers.Inputs;

public sealed record ProviderInvoiceImportRequestInput(
    string ProviderId,
    string ProcessingMonthId,
    string StorageBucket,
    string StorageObjectKey,
    string? OriginalFileName)
{
    public ProviderInvoiceImportRequestCommand ToCommand(string organizationId)
        => new(organizationId, ProviderId, ProcessingMonthId, StorageBucket, StorageObjectKey, OriginalFileName);
}