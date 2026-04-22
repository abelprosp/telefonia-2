using Goal.Infra.Data;
using Luxus.Connect.Domain.BillingCycles.Aggregates;
using Luxus.Connect.Domain.Controllership.Aggregates;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Infra.Data;

internal sealed class AppUnitOfWork(
    AppDbContext context,
    IProviderRepository providers,
    IProviderPlanRepository providerPlans,
    IProviderPlanServiceRepository providerPlanServices,
    IBillingCycleRepository billingCycles,
    IProcessingMonthRepository processingMonths,
    ICostCenterRepository costCenters,
    IContractingCompanyRepository contractingCompanies,
    ICustomerRepository customers,
    ICustomerAttachmentRepository customerAttachments,
    ICustomerProcessingMonthManualReleaseRepository customerProcessingMonthManualReleases,
    IProviderAccountRepository accounts,
    IPhoneLineRepository planLines,
    IPhoneLineServiceRepository lineServices,
    IProviderInvoiceImportRequestRepository invoiceImportRequests,
    IProviderInvoiceRepository invoices)
    : UnitOfWork(context), IAppUnitOfWork
{
    public IProviderRepository Providers { get; } = providers;
    public IProviderPlanRepository ProviderPlans { get; } = providerPlans;
    public IProviderPlanServiceRepository PlanServices { get; } = providerPlanServices;
    public IBillingCycleRepository BillingCycles { get; } = billingCycles;
    public IProcessingMonthRepository ProcessingMonths { get; } = processingMonths;
    public ICostCenterRepository CostCenters { get; } = costCenters;
    public IContractingCompanyRepository ContractingCompanies { get; } = contractingCompanies;
    public ICustomerRepository Customers { get; } = customers;
    public ICustomerAttachmentRepository CustomerAttachments { get; } = customerAttachments;
    public ICustomerProcessingMonthManualReleaseRepository CustomerProcessingMonthManualReleases { get; } =
        customerProcessingMonthManualReleases;
    public IProviderAccountRepository ProviderAccounts { get; } = accounts;
    public IPhoneLineRepository PhoneLines { get; } = planLines;
    public IPhoneLineServiceRepository LineServices { get; } = lineServices;
    public IProviderInvoiceImportRequestRepository InvoiceImportRequests { get; } = invoiceImportRequests;
    public IProviderInvoiceRepository Invoices { get; } = invoices;
}
