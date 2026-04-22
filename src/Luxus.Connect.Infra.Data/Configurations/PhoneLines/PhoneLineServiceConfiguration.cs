using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.PhoneLines;

internal sealed class PhoneLineServiceConfiguration : IEntityTypeConfiguration<PhoneLineService>
{
    public void Configure(EntityTypeBuilder<PhoneLineService> builder)
    {
        builder.ToTable("PhoneLineServices");

        builder.HasKey(x => x.Id);

        builder.Property(x => x.Id).HasMaxLength(36).IsRequired();
        builder.Property(x => x.PhoneLineId).HasMaxLength(36).IsRequired();
        builder.Property(x => x.ProviderPlanServiceId).HasMaxLength(36).IsRequired();
        builder.Property(x => x.Price).HasPrecision(18, 2);

        builder.HasIndex(x => new { x.PhoneLineId, x.ProviderPlanServiceId }).IsUnique();
    }
}
