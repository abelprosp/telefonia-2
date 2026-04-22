using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Providers;

internal sealed class ProviderInvoiceImportRequestConfiguration : IEntityTypeConfiguration<ProviderInvoiceImportRequest>
{
    public void Configure(EntityTypeBuilder<ProviderInvoiceImportRequest> builder)
    {
        builder.ToTable("ProviderInvoiceImportRequests");

        builder.HasKey(x => x.Id);

        builder.Property(x => x.Id).HasMaxLength(36).IsRequired();
        builder.Property(x => x.OrganizationId).HasMaxLength(36).IsRequired();
        builder.Property(x => x.ProviderId).HasMaxLength(36).IsRequired();
        builder.Property(x => x.ProcessingMonthId).HasMaxLength(36).IsRequired();
        builder.Property(x => x.StorageBucket).HasMaxLength(256).IsRequired();
        builder.Property(x => x.StorageObjectKey).HasMaxLength(2048).IsRequired();
        builder.Property(x => x.OriginalFileName).HasMaxLength(512);
        builder.Property(x => x.Error).HasMaxLength(8000);
        builder.Property(x => x.CreatedBy).HasMaxLength(64);

        builder.HasIndex(c => c.OrganizationId);
        builder.HasIndex(b => new { b.OrganizationId, b.ProviderId });
        builder.HasIndex(b => b.ProcessingMonthId);
    }
}
