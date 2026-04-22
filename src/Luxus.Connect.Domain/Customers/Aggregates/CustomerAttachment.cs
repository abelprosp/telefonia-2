using Goal.Domain.Aggregates;

namespace Luxus.Connect.Domain.Customers.Aggregates;

public class CustomerAttachment : Entity
{
    protected CustomerAttachment()
        : base()
    {
    }

    private CustomerAttachment(
        Customer customer,
        string? title,
        string originalFileName,
        string storageBucket,
        string storageObjectKey,
        string? contentType,
        long? sizeBytes)
        : this()
    {
        Customer = customer;
        CustomerId = customer.Id;
        OrganizationId = customer.OrganizationId;
        Title = title;
        OriginalFileName = originalFileName;
        StorageBucket = storageBucket;
        StorageObjectKey = storageObjectKey;
        ContentType = contentType;
        SizeBytes = sizeBytes;
        UploadedAtUtc = DateTimeOffset.UtcNow;
    }

    public string CustomerId { get; private set; } = default!;
    public string OrganizationId { get; private set; } = default!;
    public string? Title { get; private set; }
    public string OriginalFileName { get; private set; } = default!;
    public string StorageBucket { get; private set; } = default!;
    public string StorageObjectKey { get; private set; } = default!;
    public string? ContentType { get; private set; }
    public long? SizeBytes { get; private set; }
    public DateTimeOffset UploadedAtUtc { get; private set; }
    public Customer Customer { get; private set; } = default!;

    public static CustomerAttachment Create(
        Customer customer,
        string? title,
        string originalFileName,
        string storageBucket,
        string storageObjectKey,
        string? contentType,
        long? sizeBytes)
    {
        ArgumentNullException.ThrowIfNull(customer);
        ArgumentException.ThrowIfNullOrWhiteSpace(originalFileName);
        ArgumentException.ThrowIfNullOrWhiteSpace(storageBucket);
        ArgumentException.ThrowIfNullOrWhiteSpace(storageObjectKey);

        return new CustomerAttachment(
            customer,
            string.IsNullOrWhiteSpace(title) ? null : title.Trim(),
            originalFileName.Trim(),
            storageBucket.Trim(),
            storageObjectKey.Trim(),
            string.IsNullOrWhiteSpace(contentType) ? null : contentType.Trim(),
            sizeBytes);
    }
}
