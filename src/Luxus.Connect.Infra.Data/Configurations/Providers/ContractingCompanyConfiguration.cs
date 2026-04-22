using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Providers;

internal sealed class ContractingCompanyConfiguration : IEntityTypeConfiguration<ContractingCompany>
{
    public void Configure(EntityTypeBuilder<ContractingCompany> builder)
    {
        builder.ToTable("ContractingCompanies");

        builder.HasKey(c => c.Id);

        builder.Property(c => c.Id).HasMaxLength(36).IsRequired();
        builder.Property(c => c.ProviderId).HasMaxLength(36).IsRequired();
        builder.Property(c => c.LegalName).HasMaxLength(512).IsRequired();
        builder.Property(c => c.TaxId).HasMaxLength(14).IsRequired();

        builder.HasIndex(c => new { c.ProviderId, c.TaxId }).IsUnique();

        builder.HasMany(c => c.ProviderAccounts)
            .WithOne(cc => cc.ContractingCompany)
            .HasForeignKey(c => c.ContractingCompanyId)
            .OnDelete(DeleteBehavior.Restrict);
    }
}
