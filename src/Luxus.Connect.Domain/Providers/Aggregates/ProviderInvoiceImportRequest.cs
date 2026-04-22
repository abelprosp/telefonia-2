using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Luxus.Connect.Domain.Providers.Enums;

namespace Luxus.Connect.Domain.Providers.Aggregates;

public class ProviderInvoiceImportRequest : Entity
{
    protected ProviderInvoiceImportRequest()
        : base()
    {
    }

    private ProviderInvoiceImportRequest(
        string providerId,
        string processingMonthId,
        string storageBucket,
        string storageObjectKey,
        string organizationId,
        string createdBy)
        : this()
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(providerId);
        ArgumentException.ThrowIfNullOrWhiteSpace(processingMonthId);
        ArgumentException.ThrowIfNullOrWhiteSpace(storageBucket);
        ArgumentException.ThrowIfNullOrWhiteSpace(storageObjectKey);

        OrganizationId = organizationId;
        ProviderId = providerId;
        ProcessingMonthId = processingMonthId;
        StorageBucket = storageBucket;
        StorageObjectKey = storageObjectKey;
        Status = ProviderInvoiceImportRequestStatus.PENDING;
        CreatedBy = createdBy;
    }

    public string OrganizationId { get; private set; } = default!;
    public string ProviderId { get; private set; } = default!;
    public string ProcessingMonthId { get; private set; } = default!;
    public string StorageBucket { get; private set; } = default!;
    public string StorageObjectKey { get; private set; } = default!;
    public string? OriginalFileName { get; private set; }
    public ProviderInvoiceImportRequestStatus Status { get; private set; }
    public string? Error { get; private set; }
    public DateTimeOffset? CompletedAt { get; private set; }
    public string CreatedBy { get; private set; } = default!;
    public Provider Provider { get; private set; } = default!;
    public ProcessingMonth ProcessingMonth { get; private set; } = default!;

    public void SetOriginalFileName(string? originalFileName)
        => OriginalFileName = originalFileName;

    public void MarkProcessing()
        => Status = ProviderInvoiceImportRequestStatus.PROCESSING;

    public void MarkCompleted()
    {
        Status = ProviderInvoiceImportRequestStatus.COMPLETED;
        CompletedAt = DateTimeOffset.UtcNow;
        Error = null;
    }

    public void MarkFailed(string message)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(message);

        Status = ProviderInvoiceImportRequestStatus.FAILED;
        Error = message;
        CompletedAt = DateTimeOffset.UtcNow;
    }

    public static ProviderInvoiceImportRequest Create(
        string providerId,
        string processingMonthId,
        string storageBucket,
        string storageObjectKey,
        string organizationId,
        string createdBy)
        => new(providerId, processingMonthId, storageBucket, storageObjectKey, organizationId, createdBy);
}
