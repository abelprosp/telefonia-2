using Goal.Infra.Data;
using Luxus.Connect.Domain.BillingCycles.Aggregates;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Luxus.Connect.Domain.Providers.Enums;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;

namespace Luxus.Connect.Infra.Data;

internal class AppDbContextFactory : DesignTimeDbContextFactory<AppDbContext>
{
    protected override AppDbContext CreateNewInstance(DbContextOptionsBuilder<AppDbContext> optionsBuilder)
    {
        optionsBuilder
            .UseNpgsql(Configuration.GetConnectionString("DefaultConnection"), x =>
            {
                x.MapEnum<CustomerType>();
                x.MapEnum<CustomerDocumentType>();
                x.MapEnum<ServiceType>();
                x.MapEnum<ServiceApplicationType>();
                x.MapEnum<ServiceAvailabilityRule>();
                x.MapEnum<ExceedanceChargeType>();
                x.MapEnum<BillingCycleStatus>();
                x.MapEnum<ProcessingMonthStatus>();
                x.MapEnum<LineClassification>();
                x.MapEnum<PhoneLineStatus>();
                x.MapEnum<TransitionSubStatus>();
                x.MapEnum<ProviderInvoiceStatus>();
                x.MapEnum<ProviderInvoiceItemType>();
                x.MapEnum<InvoiceItemUnit>();
            });

        return new AppDbContext(optionsBuilder.Options);
    }
}
