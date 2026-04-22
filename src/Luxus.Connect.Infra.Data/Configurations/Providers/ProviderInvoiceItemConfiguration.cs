using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Providers;

internal sealed class ProviderInvoiceItemConfiguration : IEntityTypeConfiguration<ProviderInvoiceItem>
{
    public void Configure(EntityTypeBuilder<ProviderInvoiceItem> builder)
    {
        builder.ToTable("ProviderInvoiceItems");

        builder.HasKey(i => i.Id);

        builder.Property(i => i.Id).HasMaxLength(36).IsRequired();
        builder.Property(i => i.InvoiceId).HasMaxLength(36).IsRequired();
        builder.Property(i => i.Description).HasMaxLength(512).IsRequired();
        builder.Property(i => i.Quantity).HasPrecision(18, 4).IsRequired();
        builder.Property(i => i.TotalPrice).HasPrecision(18, 2).IsRequired();
        builder.Property(i => i.ItemType).IsRequired();
        builder.Property(i => i.QuotaAmount).HasPrecision(18, 4);
        builder.Property(i => i.ConsumedAmount).HasPrecision(18, 4);

        builder.HasMany(i => i.Children)
            .WithOne(p => p.Parent)
            .HasForeignKey(p => p.ParentId)
            .OnDelete(DeleteBehavior.Cascade);
    }
}
