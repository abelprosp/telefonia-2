using Luxus.Connect.Contracts.Customers.Responses;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Infra.Crosscutting.Extensions;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data.Query.Repositories.Customers;

internal sealed class CustomerProcessingMonthBillingReadinessQueryRepository(AppDbContext context)
    : AppQueryRepository(context), ICustomerProcessingMonthBillingReadinessQueryRepository
{
    public async Task<GetCustomerProcessingMonthBillingReadinessResponse?> LoadAsync(
        string organizationId,
        string customerId,
        string processingMonthId,
        CancellationToken cancellationToken = default)
    {
        Customer? customer = await context
            .Set<Customer>()
            .AsNoTracking()
            .Include(c => c.Documents)
            .Include(c => c.ProviderLinks)
            .SingleOrDefaultAsync(
                c => c.OrganizationId == organizationId && c.Id == customerId,
                cancellationToken);

        ProcessingMonth? month = await context
            .Set<ProcessingMonth>()
            .AsNoTracking()
            .SingleOrDefaultAsync(
                m => m.OrganizationId == organizationId && m.Id == processingMonthId,
                cancellationToken);

        if (customer is null || month is null)
            return null;

        if (!customer.HasActiveProvider(month.ProviderId))
            return null;

        CustomerProcessingMonthManualRelease? manual = await context
            .Set<CustomerProcessingMonthManualRelease>()
            .AsNoTracking()
            .SingleOrDefaultAsync(
                r => r.OrganizationId == organizationId
                    && r.CustomerId == customerId
                    && r.ProcessingMonthId == processingMonthId,
                cancellationToken);

        string? cnpjDigits = customer.Documents
            .Where(d => d.DocumentType == CustomerDocumentType.CNPJ)
            .Select(d => d.Number.NormalizeDigitsOnly())
            .FirstOrDefault(d => d.Length == 14);

        bool usesCnpjRule = !string.IsNullOrEmpty(cnpjDigits);
        int accountsExpected = 0;
        int accountsWithInvoice = 0;
        bool automaticComplete = false;

        if (usesCnpjRule && cnpjDigits is not null)
        {
            List<string> companyIds = await context
                .Set<ContractingCompany>()
                .AsNoTracking()
                .Where(cc => cc.ProviderId == month.ProviderId && cc.TaxId == cnpjDigits)
                .Select(cc => cc.Id)
                .ToListAsync(cancellationToken);

            List<string> accountIds = await context
                .Set<ProviderAccount>()
                .AsNoTracking()
                .Where(a => companyIds.Contains(a.ContractingCompanyId))
                .Select(a => a.Id)
                .ToListAsync(cancellationToken);

            accountsExpected = accountIds.Count;

            if (accountsExpected > 0)
            {
                accountsWithInvoice = await context
                    .Set<ProviderInvoice>()
                    .AsNoTracking()
                    .Where(i =>
                        i.ProcessingMonthId == processingMonthId
                        && accountIds.Contains(i.ProviderAccountId))
                    .Select(i => i.ProviderAccountId)
                    .Distinct()
                    .CountAsync(cancellationToken);

                automaticComplete = accountsWithInvoice == accountsExpected;
            }
        }

        bool isManuallyReleased = manual is not null;
        bool isReleased = isManuallyReleased || automaticComplete;

        string statusName = isReleased
            ? "Liberado para faturamento"
            : "Pendente";

        GetCustomerProcessingMonthBillingReadinessManualReleaseDto? manualDto = manual is null
            ? null
            : new GetCustomerProcessingMonthBillingReadinessManualReleaseDto(
                manual.Justification,
                manual.ReleasedAt,
                manual.ReleasedByUserId);

        return new GetCustomerProcessingMonthBillingReadinessResponse(
            customer.Id,
            month.Id,
            statusName,
            isReleased,
            automaticComplete,
            isManuallyReleased,
            usesCnpjRule,
            accountsExpected,
            accountsWithInvoice,
            manualDto);
    }
}
