using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Providers;

internal sealed class ProviderInvoiceServiceConfiguration : IEntityTypeConfiguration<ProviderInvoiceService>
{
    public void Configure(EntityTypeBuilder<ProviderInvoiceService> builder)
    {
        builder.ToTable("ProviderInvoiceServices");

        builder.HasKey(s => s.Id);

        builder.Property(s => s.Id).HasMaxLength(36).IsRequired();
        builder.Property(s => s.InvoiceId).HasMaxLength(36).IsRequired();
        builder.Property(s => s.PlanId).HasMaxLength(36);
        builder.Property(i => i.Description).HasMaxLength(512).IsRequired();
        builder.Property(i => i.Quantity).HasPrecision(18, 4).IsRequired();
        builder.Property(i => i.TotalPrice).HasPrecision(18, 2).IsRequired();
        builder.Property(i => i.QuotaAmount).HasPrecision(18, 4);
        builder.Property(i => i.ConsumedAmount).HasPrecision(18, 4);
    }
}
