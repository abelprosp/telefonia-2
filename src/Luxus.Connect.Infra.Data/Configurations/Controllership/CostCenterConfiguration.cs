using Luxus.Connect.Domain.Controllership.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Controllership;

internal sealed class CostCenterConfiguration : IEntityTypeConfiguration<CostCenter>
{
    public void Configure(EntityTypeBuilder<CostCenter> builder)
    {
        builder.ToTable("CostCenters");

        builder.HasKey(x => x.Id);

        builder.Property(x => x.Id).HasMaxLength(36).IsRequired();
        builder.Property(x => x.OrganizationId).HasMaxLength(36).IsRequired();
        builder.Property(x => x.Name).HasMaxLength(128).IsRequired();
        builder.Property(x => x.Description).HasMaxLength(256).IsRequired();

        builder.HasIndex(c => c.OrganizationId);

        builder.HasIndex(b => new { b.OrganizationId, b.Name }).IsUnique();

        builder.HasMany(p => p.PhoneLines)
            .WithOne(q => q.CostCenter)
            .HasForeignKey(q => q.CostCenterId)
            .OnDelete(DeleteBehavior.SetNull);

        builder.HasMany(p => p.ProviderInvoices)
            .WithOne(q => q.CostCenter)
            .HasForeignKey(q => q.CostCenterId)
            .OnDelete(DeleteBehavior.SetNull);
    }
}
