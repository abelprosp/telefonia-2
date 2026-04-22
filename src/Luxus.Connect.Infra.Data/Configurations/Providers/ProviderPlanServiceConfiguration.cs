using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Providers;

internal sealed class ProviderPlanServiceConfiguration : IEntityTypeConfiguration<ProviderPlanService>
{
    public void Configure(EntityTypeBuilder<ProviderPlanService> builder)
    {
        builder.ToTable("ProviderPlanServices");

        builder.HasKey(x => x.Id);

        builder.Property(x => x.Id).HasMaxLength(36).IsRequired();
        builder.Property(x => x.ProviderPlanId).HasMaxLength(36).IsRequired();
        builder.Property(x => x.Name).HasMaxLength(256).IsRequired();
        builder.Property(x => x.Active).HasDefaultValue(true);
        builder.Property(x => x.Price).HasPrecision(18, 2);
    }
}
