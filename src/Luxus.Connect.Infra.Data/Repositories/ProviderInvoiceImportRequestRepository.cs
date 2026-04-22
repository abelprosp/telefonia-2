using Goal.Infra.Data;
using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Infra.Data.Repositories;

internal sealed class ProviderInvoiceImportRequestRepository(AppDbContext context)
    : Repository<ProviderInvoiceImportRequest>(context), IProviderInvoiceImportRequestRepository;
