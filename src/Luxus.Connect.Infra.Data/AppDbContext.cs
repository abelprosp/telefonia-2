using Goal.Infra.Data.Auditing;
using Luxus.Connect.Domain.BillingCycles.Aggregates;
using Luxus.Connect.Domain.Controllership.Aggregates;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Infra.Data.Configurations.Auditing;
using Luxus.Connect.Infra.Data.Configurations.BillingCycles;
using Luxus.Connect.Infra.Data.Configurations.Controllership;
using Luxus.Connect.Infra.Data.Configurations.Customers;
using Luxus.Connect.Infra.Data.Configurations.PhoneLines;
using Luxus.Connect.Infra.Data.Configurations.ProcessingMonths;
using Luxus.Connect.Infra.Data.Configurations.Providers;
using Microsoft.EntityFrameworkCore;

namespace Luxus.Connect.Infra.Data;

public sealed class AppDbContext(DbContextOptions<AppDbContext> options)
    : DbContext(options)
{
    public DbSet<AuditLog> AuditLogs { get; set; }
    public DbSet<Provider> Providers { get; set; }
    public DbSet<ProviderPlan> ProviderPlans { get; set; }
    public DbSet<ProviderPlanService> ProviderPlanServices { get; set; }
    public DbSet<PhoneLine> PhoneLines { get; set; }
    public DbSet<PhoneLineCustomerLink> PhoneLineCustomerLinks { get; set; }
    public DbSet<PhoneLineService> PhoneLineService { get; set; }
    public DbSet<BillingCycle> BillingCycles { get; set; }
    public DbSet<ProcessingMonth> ProcessingMonths { get; set; }
    public DbSet<CostCenter> CostCenters { get; set; }
    public DbSet<ContractingCompany> ContractingCompanies { get; set; }
    public DbSet<Customer> Customers { get; set; }
    public DbSet<CustomerProviderLink> CustomerProviderLinks { get; set; }
    public DbSet<CustomerProcessingMonthManualRelease> CustomerProcessingMonthManualReleases { get; set; }
    public DbSet<CustomerAddress> CustomerAddresses { get; set; }
    public DbSet<CustomerDocument> CustomerDocuments { get; set; }
    public DbSet<CustomerAttachment> CustomerAttachments { get; set; }
    public DbSet<ProviderAccount> ProviderAccounts { get; set; }
    public DbSet<ProviderInvoiceImportRequest> ProviderInvoiceImportRequests { get; set; }
    public DbSet<ProviderInvoice> ProviderInvoices { get; set; }
    public DbSet<ProviderInvoiceItem> ProviderInvoiceItems { get; set; }
    public DbSet<ProviderInvoiceService> ProviderInvoiceServices { get; set; }
    public DbSet<ProviderInvoiceQuotaSharing> ProviderInvoiceQuotaSharing { get; set; }

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder
            .ApplyConfiguration(new AuditLogConfiguration())
            .ApplyConfiguration(new BillingCycleConfiguration())
            .ApplyConfiguration(new ProcessingMonthConfiguration())
            .ApplyConfiguration(new CostCenterConfiguration())
            .ApplyConfiguration(new CustomerConfiguration())
            .ApplyConfiguration(new CustomerProviderLinkConfiguration())
            .ApplyConfiguration(new CustomerProcessingMonthManualReleaseConfiguration())
            .ApplyConfiguration(new CustomerAddressConfiguration())
            .ApplyConfiguration(new CustomerDocumentConfiguration())
            .ApplyConfiguration(new CustomerAttachmentConfiguration())
            .ApplyConfiguration(new PhoneLineConfiguration())
            .ApplyConfiguration(new PhoneLineCustomerLinkConfiguration())
            .ApplyConfiguration(new PhoneLineServiceConfiguration())
            .ApplyConfiguration(new ContractingCompanyConfiguration())
            .ApplyConfiguration(new ProviderAccountConfiguration())
            .ApplyConfiguration(new ProviderConfiguration())
            .ApplyConfiguration(new ProviderInvoiceConfiguration())
            .ApplyConfiguration(new ProviderInvoiceImportRequestConfiguration())
            .ApplyConfiguration(new ProviderInvoiceItemConfiguration())
            .ApplyConfiguration(new ProviderInvoiceQuotaSharingConfiguration())
            .ApplyConfiguration(new ProviderInvoiceServiceConfiguration())
            .ApplyConfiguration(new ProviderPlanConfiguration())
            .ApplyConfiguration(new ProviderPlanServiceConfiguration());
    }
}
