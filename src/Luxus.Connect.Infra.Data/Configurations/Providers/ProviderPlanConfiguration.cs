using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Providers;

internal sealed class ProviderPlanConfiguration : IEntityTypeConfiguration<ProviderPlan>
{
    public void Configure(EntityTypeBuilder<ProviderPlan> builder)
    {
        builder.ToTable("ProviderPlans");

        builder.HasKey(p => p.Id);

        builder.Property(p => p.Id).HasMaxLength(36).IsRequired();
        builder.Property(p => p.ProviderId).HasMaxLength(36).IsRequired();
        builder.Property(p => p.Name).HasMaxLength(256).IsRequired();
        builder.Property(p => p.Code).HasMaxLength(64).IsRequired();

        builder.HasIndex(p => new { p.ProviderId, p.Code }).IsUnique();

        builder.HasMany(p => p.ProviderPlanServices)
            .WithOne(x => x.ProviderPlan)
            .HasForeignKey(x => x.ProviderPlanId)
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasMany(p => p.PhoneLines)
            .WithOne(l => l.ProviderPlan)
            .HasForeignKey(l => l.ProviderPlanId)
            .OnDelete(DeleteBehavior.SetNull);
    }
}
