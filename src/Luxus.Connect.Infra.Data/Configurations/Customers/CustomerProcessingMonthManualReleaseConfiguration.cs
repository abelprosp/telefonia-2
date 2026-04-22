using Luxus.Connect.Domain.Customers.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Customers;

internal sealed class CustomerProcessingMonthManualReleaseConfiguration
    : IEntityTypeConfiguration<CustomerProcessingMonthManualRelease>
{
    public void Configure(EntityTypeBuilder<CustomerProcessingMonthManualRelease> builder)
    {
        builder.ToTable("CustomerProcessingMonthManualReleases");

        builder.HasKey(e => e.Id);

        builder.Property(e => e.Id).HasMaxLength(36).IsRequired();
        builder.Property(e => e.OrganizationId).HasMaxLength(36).IsRequired();
        builder.Property(e => e.CustomerId).HasMaxLength(36).IsRequired();
        builder.Property(e => e.ProcessingMonthId).HasMaxLength(36).IsRequired();
        builder.Property(e => e.Justification).HasMaxLength(4000).IsRequired();
        builder.Property(e => e.ReleasedByUserId).HasMaxLength(64).IsRequired();
        builder.Property(e => e.ReleasedAt).IsRequired();

        builder.HasIndex(e => e.OrganizationId);

        builder.HasIndex(e => new { e.CustomerId, e.ProcessingMonthId })
            .IsUnique();
    }
}
