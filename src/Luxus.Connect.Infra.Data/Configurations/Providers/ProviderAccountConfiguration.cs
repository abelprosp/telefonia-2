using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Providers;

internal sealed class ProviderAccountConfiguration : IEntityTypeConfiguration<ProviderAccount>
{
    public void Configure(EntityTypeBuilder<ProviderAccount> builder)
    {
        builder.ToTable("ProviderAccounts");

        builder.HasKey(a => a.Id);

        builder.Property(a => a.Id).HasMaxLength(36).IsRequired();
        builder.Property(a => a.ContractingCompanyId).HasMaxLength(36).IsRequired();
        builder.Property(a => a.AccountNumber).HasMaxLength(64).IsRequired();

        builder.HasIndex(a => a.AccountNumber);
        builder.HasIndex(a => new { a.ContractingCompanyId, a.AccountNumber }).IsUnique();

        builder.HasMany(i => i.BillingCycles)
            .WithMany(l => l.ProviderAccounts)
            .UsingEntity(j => j.ToTable("BillingCycleProviderAccounts"));
    }
}
