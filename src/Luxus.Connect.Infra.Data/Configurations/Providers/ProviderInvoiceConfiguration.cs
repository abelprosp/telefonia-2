using Luxus.Connect.Domain.Providers.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Providers;

internal sealed class ProviderInvoiceConfiguration : IEntityTypeConfiguration<ProviderInvoice>
{
    public void Configure(EntityTypeBuilder<ProviderInvoice> builder)
    {
        builder.ToTable("ProviderInvoices");

        builder.HasKey(i => i.Id);

        builder.Property(i => i.Id).HasMaxLength(36).IsRequired();
        builder.Property(i => i.ProviderAccountId).HasMaxLength(36).IsRequired();
        builder.Property(i => i.ContractingCompanyId).HasMaxLength(36).IsRequired();
        builder.Property(i => i.ProcessingMonthId).HasMaxLength(36).IsRequired();
        builder.Property(i => i.BillingCycleId).HasMaxLength(36).IsRequired();
        builder.Property(i => i.CostCenterId).HasMaxLength(36);
        builder.Property(i => i.ParentInvoiceId).HasMaxLength(36);
        builder.Property(i => i.IssueDate).IsRequired();
        builder.Property(i => i.DueDate).IsRequired();
        builder.Property(i => i.TotalAmount).HasPrecision(18, 2).IsRequired();
        builder.Property(i => i.SubtotalServices).HasPrecision(18, 2);
        builder.Property(i => i.SubtotalUsage).HasPrecision(18, 2);
        builder.Property(i => i.SubtotalTaxes).HasPrecision(18, 2);
        builder.Property(i => i.SubtotalDiscounts).HasPrecision(18, 2);
        builder.Property(i => i.SubtotalInstallments).HasPrecision(18, 2);
        builder
            .HasIndex(i => new
            {
                i.ProviderAccountId,
                i.ContractingCompanyId,
                i.ProcessingMonthId,
                i.DueDate
            })
            .IsUnique();
        builder.HasIndex(i => i.ProcessingMonthId);

        builder.HasMany(i => i.ProviderInvoiceItems)
            .WithOne(it => it.Invoice)
            .HasForeignKey(it => it.InvoiceId)
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasMany(i => i.ProviderInvoiceServices)
            .WithOne(s => s.Invoice)
            .HasForeignKey(s => s.InvoiceId)
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasMany(i => i.ProviderInvoiceQuotaSharing)
            .WithOne(q => q.Invoice)
            .HasForeignKey(q => q.InvoiceId)
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasMany(i => i.LastPhoneLines)
            .WithOne(q => q.LastInvoice)
            .HasForeignKey(q => q.LastInvoiceId)
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasMany(i => i.PhoneLines)
            .WithMany(l => l.ProviderInvoices)
            .UsingEntity(j => j.ToTable("ProviderInvoicePhoneLines"));
    }
}
