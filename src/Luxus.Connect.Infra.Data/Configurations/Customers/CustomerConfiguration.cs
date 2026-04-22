using Luxus.Connect.Domain.Customers.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Customers;

internal sealed class CustomerConfiguration : IEntityTypeConfiguration<Customer>
{
    public void Configure(EntityTypeBuilder<Customer> builder)
    {
        builder.ToTable("Customers");

        builder.HasKey(c => c.Id);

        builder.Property(c => c.Id).HasMaxLength(36).IsRequired();
        builder.Property(c => c.OrganizationId).HasMaxLength(36).IsRequired();
        builder.Property(c => c.Name).HasMaxLength(256).IsRequired();
        builder.Property(c => c.LegalName).HasMaxLength(512);
        builder.Property(c => c.ResponsibleSalespersonUserId).HasMaxLength(256);

        builder.HasIndex(c => c.OrganizationId);
        builder.HasIndex(b => new { b.OrganizationId, b.Name });
        builder.HasIndex(b => new { b.OrganizationId, b.LegalName }).IsUnique();

        builder.HasMany(c => c.Addresses)
            .WithOne(a => a.Customer)
            .HasForeignKey(a => a.CustomerId)
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasMany(c => c.Documents)
            .WithOne(d => d.Customer)
            .HasForeignKey(d => d.CustomerId)
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasMany(c => c.ProviderLinks)
            .WithOne(d => d.Customer)
            .HasForeignKey(d => d.CustomerId)
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasMany(c => c.CustomerProcessingMonthManualReleases)
            .WithOne(d => d.Customer)
            .HasForeignKey(d => d.CustomerId)
            .OnDelete(DeleteBehavior.Cascade);
    }
}
