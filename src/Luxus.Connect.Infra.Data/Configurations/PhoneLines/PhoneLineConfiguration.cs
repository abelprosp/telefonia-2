using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.PhoneLines;

internal sealed class PhoneLineConfiguration : IEntityTypeConfiguration<PhoneLine>
{
    public void Configure(EntityTypeBuilder<PhoneLine> builder)
    {
        builder.ToTable("PhoneLines");

        builder.HasKey(p => p.Id);

        builder.Property(p => p.Id).HasMaxLength(36).IsRequired();
        builder.Property(p => p.ProviderPlanId).HasMaxLength(36).IsRequired();
        builder.Property(p => p.ProviderAccountId).HasMaxLength(36).IsRequired();
        builder.Property(p => p.CostCenterId).HasMaxLength(36);
        builder.Property(p => p.LastInvoiceId).HasMaxLength(36);
        builder.Property(p => p.TitularLineId).HasMaxLength(36);
        builder.Property(p => p.Number).HasMaxLength(20);
        builder.Property(p => p.BaseCost).HasPrecision(18, 2);
        builder.Property(p => p.CostWithConsumption).HasPrecision(18, 2);

        builder.HasIndex(p => p.Number).IsUnique();
        builder.HasIndex(p => new { p.Status, p.TransitionStartedAt });

        builder.HasMany(p => p.ChildrenLines)
            .WithOne(c => c.TitularLine)
            .HasForeignKey(c => c.TitularLineId)
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasMany(p => p.InvoiceQuotaSharings)
            .WithOne(q => q.PhoneLine)
            .HasForeignKey(q => q.PhoneLineId)
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasMany(p => p.PhoneLineServices)
            .WithOne(q => q.PhoneLine)
            .HasForeignKey(q => q.PhoneLineId)
            .OnDelete(DeleteBehavior.Cascade);
    }
}
