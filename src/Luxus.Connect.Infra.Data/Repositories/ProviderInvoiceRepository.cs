using Goal.Infra.Data;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Domain.Providers.Enums;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Repositories;

internal sealed class ProviderInvoiceRepository(AppDbContext context)
    : Repository<ProviderInvoice>(context), IProviderInvoiceRepository
{
    public Task<bool> ExistsInCycleAsync(
        string cycleId,
        string accountNumber,
        CancellationToken cancellationToken = default)
    {
        return Context
            .Set<ProviderInvoice>()
            .AnyAsync(
                i => i.BillingCycleId == cycleId && i.ProviderAccount.AccountNumber == accountNumber,
                cancellationToken);
    }

    public async Task<ProviderInvoice?> GetWithDetailsAsync(string id, CancellationToken cancellationToken = default)
    {
        return await Context
            .Set<ProviderInvoice>()
            .Include(i => i.ProviderInvoiceItems)
            .Include(i => i.ProviderInvoiceServices)
            .Include(i => i.ProviderInvoiceQuotaSharing)
            .Include(i => i.PhoneLines)
            .Include(i => i.BillingCycle)
            .Include(i => i.ContractingCompany)
            .SingleOrDefaultAsync(i => i.Id == id, cancellationToken);
    }

    public async Task<ProviderInvoiceDuplication> FindDuplicateByBusinessKeyAsync(
        string accountNumber,
        string contractingCompanyId,
        string processingMonthId,
        DateOnly dueDate,
        CancellationToken cancellationToken = default)
    {
        bool exists = await Context
            .Set<ProviderInvoice>()
            .AsNoTracking()
            .AnyAsync(
                i =>
                    i.ProviderAccount.AccountNumber == accountNumber
                    && i.ContractingCompanyId == contractingCompanyId
                    && i.ProcessingMonthId == processingMonthId
                    && i.DueDate == dueDate,
                cancellationToken);

        return exists
            ? ProviderInvoiceDuplication.Duplicate
            : ProviderInvoiceDuplication.None;
    }

    public Task<bool> IsPhoneLineLinkedToAnyInvoiceInBillingCycleAsync(
        string phoneLineId,
        string processingMonthId,
        CancellationToken cancellationToken = default)
    {
        return Context
            .Set<ProviderInvoice>()
            .AnyAsync(
                i => i.ProcessingMonthId == processingMonthId
                && i.PhoneLines.Any(l => l.Id == phoneLineId),
                cancellationToken);
    }
}
