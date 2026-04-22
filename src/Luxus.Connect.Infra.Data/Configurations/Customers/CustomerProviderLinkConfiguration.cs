using Luxus.Connect.Domain.Customers.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Customers;

internal sealed class CustomerProviderLinkConfiguration : IEntityTypeConfiguration<CustomerProviderLink>
{
    public void Configure(EntityTypeBuilder<CustomerProviderLink> builder)
    {
        builder.ToTable("CustomerProviderLinks");

        builder.HasKey(c => c.Id);

        builder.Property(c => c.Id).HasMaxLength(36).IsRequired();
        builder.Property(c => c.CustomerId).HasMaxLength(36).IsRequired();
        builder.Property(c => c.ProviderId).HasMaxLength(36).IsRequired();
        builder.Property(c => c.StartDate).IsRequired();
        builder.Property(c => c.EndDate);

        builder.HasIndex(c => new { c.CustomerId, c.ProviderId, c.StartDate });
        builder.HasIndex(c => new { c.CustomerId, c.ProviderId, c.EndDate });
        builder.HasIndex(c => new { c.CustomerId, c.ProviderId })
            .HasFilter("\"EndDate\" IS NULL")
            .IsUnique();
    }
}
