using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.ProcessingMonths;

internal sealed class ProcessingMonthConfiguration : IEntityTypeConfiguration<ProcessingMonth>
{
    public void Configure(EntityTypeBuilder<ProcessingMonth> builder)
    {
        builder.ToTable("ProcessingMonths");

        builder.HasKey(p => p.Id);

        builder.Property(p => p.Id).HasMaxLength(36).IsRequired();
        builder.Property(p => p.OrganizationId).HasMaxLength(36).IsRequired();
        builder.Property(p => p.ProviderId).HasMaxLength(36).IsRequired();
        builder.Property(p => p.Year).IsRequired();
        builder.Property(p => p.Month).IsRequired();
        builder.Property(p => p.DisplayName).HasMaxLength(128).IsRequired();
        builder.Property(p => p.Status).IsRequired();
        builder.Property(p => p.ClosedAt);
        builder.Property(p => p.ClosedBy).HasMaxLength(36);
        builder.Property(p => p.ClosedInContingency).IsRequired();
        builder.Property(p => p.ContingencyJustification).HasMaxLength(4000);

        builder.HasIndex(p => p.OrganizationId);

        builder.HasIndex(p => new { p.OrganizationId, p.ProviderId, p.Year, p.Month })
            .IsUnique();

        builder.HasMany(c => c.CustomerProcessingMonthManualReleases)
            .WithOne(d => d.ProcessingMonth)
            .HasForeignKey(d => d.ProcessingMonthId)
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasMany(c => c.ProviderInvoices)
            .WithOne(d => d.ProcessingMonth)
            .HasForeignKey(d => d.ProcessingMonthId)
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasMany(c => c.ProviderInvoiceImportRequests)
            .WithOne(d => d.ProcessingMonth)
            .HasForeignKey(d => d.ProcessingMonthId)
            .OnDelete(DeleteBehavior.Cascade);
    }
}
