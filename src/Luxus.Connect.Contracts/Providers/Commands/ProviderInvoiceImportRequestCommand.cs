using Goal.Application.Commands;
using Luxus.Connect.Contracts.Providers.Responses;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;

namespace Luxus.Connect.Contracts.Providers.Commands;

public sealed record ProviderInvoiceImportRequestCommand(
    string OrganizationId,
    string ProviderId,
    string ProcessingMonthId,
    string StorageBucket,
    string StorageObjectKey,
    string? OriginalFileName = null
) : ICommand<OneOf<RequestProviderInvoiceImportResponse, AppError>>;
