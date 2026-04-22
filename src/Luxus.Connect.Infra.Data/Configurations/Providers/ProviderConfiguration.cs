using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Providers;

internal sealed class ProviderConfiguration : IEntityTypeConfiguration<Provider>
{
    public void Configure(EntityTypeBuilder<Provider> builder)
    {
        builder.ToTable("Providers");

        builder.HasKey(o => o.Id);

        builder.Property(o => o.Id)
            .HasMaxLength(36)
            .IsRequired();

        builder.Property(o => o.OrganizationId)
            .HasMaxLength(36)
            .IsRequired();

        builder.Property(o => o.Name)
            .HasMaxLength(256)
            .IsRequired();

        builder.Property(o => o.Slug)
            .HasMaxLength(256)
            .IsRequired();

        builder.Property(o => o.Active)
            .HasDefaultValue(true);

        builder.HasIndex(c => c.OrganizationId);
        builder.HasIndex(b => new { b.OrganizationId, b.Active });
        builder.HasIndex(b => new { b.OrganizationId, b.Slug }).IsUnique();

        builder.HasMany(o => o.ProviderPlans)
            .WithOne(p => p.Provider)
            .HasForeignKey(p => p.ProviderId)
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasMany(o => o.ContractingCompanies)
            .WithOne(l => l.Provider)
            .HasForeignKey(l => l.ProviderId)
            .OnDelete(DeleteBehavior.Restrict);

        builder.HasMany(o => o.InvoiceImportRequests)
            .WithOne(l => l.Provider)
            .HasForeignKey(l => l.ProviderId)
            .OnDelete(DeleteBehavior.Restrict);

        builder.HasMany(p => p.ProcessingMonths)
            .WithOne(p => p.Provider)
            .HasForeignKey(p => p.ProviderId)
            .OnDelete(DeleteBehavior.Restrict);

        builder.HasMany(c => c.ProviderLinks)
            .WithOne(d => d.Provider)
            .HasForeignKey(d => d.ProviderId)
            .OnDelete(DeleteBehavior.Cascade);
    }
}
