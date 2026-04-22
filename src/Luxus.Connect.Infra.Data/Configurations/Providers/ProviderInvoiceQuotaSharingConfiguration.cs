using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Providers;

internal sealed class ProviderInvoiceQuotaSharingConfiguration : IEntityTypeConfiguration<ProviderInvoiceQuotaSharing>
{
    public void Configure(EntityTypeBuilder<ProviderInvoiceQuotaSharing> builder)
    {
        builder.ToTable("ProviderInvoiceQuotaSharing");

        builder.HasKey(q => q.Id);

        builder.Property(q => q.Id).HasMaxLength(36).IsRequired();
        builder.Property(q => q.InvoiceId).HasMaxLength(36).IsRequired();
        builder.Property(q => q.PhoneLineId).HasMaxLength(36).IsRequired();
        builder.Property(i => i.Description).HasMaxLength(512).IsRequired();
        builder.Property(i => i.ConsumedAmount).HasPrecision(18, 4);
    }
}
