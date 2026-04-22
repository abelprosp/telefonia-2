using Luxus.Connect.Domain.Customers.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Customers;

internal sealed class CustomerAddressConfiguration : IEntityTypeConfiguration<CustomerAddress>
{
    public void Configure(EntityTypeBuilder<CustomerAddress> builder)
    {
        builder.ToTable("CustomerAddresses");

        builder.HasKey(c => c.Id);

        builder.Property(c => c.Id).HasMaxLength(36).IsRequired();
        builder.Property(c => c.CustomerId).HasMaxLength(36).IsRequired();
        builder.Property(c => c.Street).HasMaxLength(256).IsRequired();
        builder.Property(c => c.Number).HasMaxLength(16).IsRequired();
        builder.Property(c => c.Complement).HasMaxLength(128);
        builder.Property(c => c.Neighborhood).HasMaxLength(128);
        builder.Property(c => c.City).HasMaxLength(128).IsRequired();
        builder.Property(c => c.State).HasMaxLength(64).IsRequired();
        builder.Property(c => c.ZipCode).HasMaxLength(10).IsRequired();
        builder.Property(c => c.Country).HasMaxLength(32).IsRequired();
    }
}
