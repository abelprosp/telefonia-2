using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.Providers.Enums;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public interface IProviderInvoiceRepository : IRepository<ProviderInvoice>
{
    Task<bool> ExistsInCycleAsync(string cycleId, string accountNumber, CancellationToken cancellationToken = default);
    Task<ProviderInvoice?> GetWithDetailsAsync(string id, CancellationToken cancellationToken = default);
    Task<ProviderInvoiceDuplication> FindDuplicateByBusinessKeyAsync(string accountNumber, string contractingCompanyId, string processingMonthId, DateOnly dueDate, CancellationToken cancellationToken = default);
    Task<bool> IsPhoneLineLinkedToAnyInvoiceInBillingCycleAsync(string phoneLineId, string processingMonthId, CancellationToken cancellationToken = default);
}
