using Luxus.Connect.Domain.Customers.Aggregates;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Customers;

internal sealed class CustomerAttachmentConfiguration : IEntityTypeConfiguration<CustomerAttachment>
{
    public void Configure(EntityTypeBuilder<CustomerAttachment> builder)
    {
        builder.ToTable("CustomerAttachments");

        builder.HasKey(a => a.Id);

        builder.Property(a => a.Id).HasMaxLength(36).IsRequired();
        builder.Property(a => a.CustomerId).HasMaxLength(36).IsRequired();
        builder.Property(a => a.OrganizationId).HasMaxLength(36).IsRequired();
        builder.Property(a => a.Title).HasMaxLength(256);
        builder.Property(a => a.OriginalFileName).HasMaxLength(512).IsRequired();
        builder.Property(a => a.StorageBucket).HasMaxLength(256).IsRequired();
        builder.Property(a => a.StorageObjectKey).HasMaxLength(2048).IsRequired();
        builder.Property(a => a.ContentType).HasMaxLength(128);
        builder.Property(a => a.UploadedAtUtc).IsRequired();

        builder.HasIndex(a => new { a.OrganizationId, a.CustomerId });

        builder
            .HasOne(a => a.Customer)
            .WithMany(p => p.Attachments)
            .HasForeignKey(a => a.CustomerId)
            .OnDelete(DeleteBehavior.Cascade);
    }
}
