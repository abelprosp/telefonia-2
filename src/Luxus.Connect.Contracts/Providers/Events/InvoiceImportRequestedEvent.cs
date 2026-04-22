using Goal.Application.Events;

namespace Luxus.Connect.Contracts.Providers.Events;

public sealed class InvoiceImportRequestedEvent(
    string aggregateId,
    string storageBucket,
    string storageObjectKey,
    string? originalFileName,
    string requestedByUserId)
    : Event(aggregateId, nameof(InvoiceImportRequestedEvent))
{
    private InvoiceImportRequestedEvent()
        : this(string.Empty, string.Empty, string.Empty, null, string.Empty)
    {
    }

    public const string RabbitMqRoutingKey = "providers.invoice_import.requested";

    public string StorageBucket { get; } = storageBucket;
    public string StorageObjectKey { get; } = storageObjectKey;
    public string? OriginalFileName { get; } = originalFileName;
    public string RequestedByUserId { get; } = requestedByUserId;
}
