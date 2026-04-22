using Luxus.Connect.Domain.BillingCycles.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.BillingCycles;

internal sealed class BillingCycleConfiguration : IEntityTypeConfiguration<BillingCycle>
{
    public void Configure(EntityTypeBuilder<BillingCycle> builder)
    {
        builder.ToTable("BillingCycles");

        builder.HasKey(b => b.Id);

        builder.Property(b => b.Id).HasMaxLength(36).IsRequired();
        builder.Property(b => b.OrganizationId).HasMaxLength(36).IsRequired();
        builder.Property(b => b.ProviderId).HasMaxLength(36).IsRequired();
        builder.Property(b => b.Code).HasMaxLength(20).IsRequired();
        builder.Property(b => b.Name).HasMaxLength(100).IsRequired();
        builder.Property(b => b.StartDate).HasConversion<DateOnly>();
        builder.Property(b => b.EndDate).HasConversion<DateOnly>();
        builder.Property(b => b.Status).IsRequired();
        builder.Property(b => b.ClosedAt);
        builder.Property(b => b.ClosedBy).HasMaxLength(36);

        builder.HasIndex(c => c.OrganizationId);

        builder.HasIndex(b => new { b.OrganizationId, b.Code }).IsUnique();
        builder.HasIndex(b => new { b.OrganizationId, b.Name }).IsUnique();

        builder.HasMany(b => b.ProviderInvoices)
            .WithOne(i => i.BillingCycle)
            .HasForeignKey(i => i.BillingCycleId)
            .OnDelete(DeleteBehavior.Restrict);
    }
}
