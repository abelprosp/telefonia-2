using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.PhoneLines;

internal sealed class PhoneLineCustomerLinkConfiguration : IEntityTypeConfiguration<PhoneLineCustomerLink>
{
    public void Configure(EntityTypeBuilder<PhoneLineCustomerLink> builder)
    {
        builder.ToTable("PhoneLineCustomerLinks");

        builder.HasKey(p => p.Id);

        builder.Property(p => p.Id).HasMaxLength(36).IsRequired();
        builder.Property(p => p.PhoneLineId).HasMaxLength(36).IsRequired();
        builder.Property(p => p.CustomerId).HasMaxLength(36).IsRequired();
        builder.Property(p => p.StartDate).IsRequired();
        builder.Property(p => p.EndDate);

        builder.HasIndex(p => new { p.PhoneLineId, p.StartDate });
        builder.HasIndex(p => p.CustomerId);
        builder.HasIndex(p => p.PhoneLineId)
            .IsUnique()
            .HasFilter("\"EndDate\" IS NULL");

        builder.HasOne(p => p.PhoneLine)
            .WithMany(l => l.CustomerLinks)
            .HasForeignKey(p => p.PhoneLineId)
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasOne(p => p.Customer)
            .WithMany(c => c.PhoneLineLinks)
            .HasForeignKey(p => p.CustomerId)
            .OnDelete(DeleteBehavior.Restrict);
    }
}
