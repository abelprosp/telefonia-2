using Goal.Domain;
using Luxus.Connect.Domain.BillingCycles.Aggregates;
using Luxus.Connect.Domain.Controllership.Aggregates;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;

namespace Luxus.Connect.Infra.Data;

public interface IAppUnitOfWork : IUnitOfWork
{
    IProviderRepository Providers { get; }
    IProviderPlanRepository ProviderPlans { get; }
    IProviderPlanServiceRepository PlanServices { get; }
    IBillingCycleRepository BillingCycles { get; }
    IProcessingMonthRepository ProcessingMonths { get; }
    ICostCenterRepository CostCenters { get; }
    IContractingCompanyRepository ContractingCompanies { get; }
    ICustomerRepository Customers { get; }
    ICustomerAttachmentRepository CustomerAttachments { get; }
    ICustomerProcessingMonthManualReleaseRepository CustomerProcessingMonthManualReleases { get; }
    IProviderAccountRepository ProviderAccounts { get; }
    IPhoneLineRepository PhoneLines { get; }
    IPhoneLineServiceRepository LineServices { get; }
    IProviderInvoiceImportRequestRepository InvoiceImportRequests { get; }
    IProviderInvoiceRepository Invoices { get; }
}
