using Goal.Infra.Data.Auditing;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Luxus.Connect.Infra.Data.Configurations.Auditing;

internal class AuditLogConfiguration : IEntityTypeConfiguration<AuditLog>
{
    public void Configure(EntityTypeBuilder<AuditLog> builder)
    {
        builder.ToTable("AuditLogs");

        builder.HasKey(b => b.Id);

        builder.Property(b => b.Id).HasMaxLength(36).IsRequired();
        builder.Property(b => b.ChangeType).HasMaxLength(16).IsRequired();
        builder.Property(b => b.EntityName).HasMaxLength(64).IsRequired();
        builder.Property(b => b.KeyValues).IsRequired();
        builder.Property(b => b.ChangedBy).HasMaxLength(36);
        builder.Property(b => b.OldValues);
        builder.Property(b => b.NewValues);
        builder.Property(b => b.Timestamp);

        builder.HasIndex(b => b.EntityName);
        builder.HasIndex(b => new { b.EntityName, b.ChangedBy });
    }
}
