using Luxus.Connect.Contracts.Customers.Commands;

namespace Luxus.Connect.Contracts.Customers.Inputs;

public sealed record RegisterCustomerAttachmentInput(
    string? Title,
    string OriginalFileName,
    string StorageBucket,
    string StorageObjectKey,
    string? ContentType = null,
    long? SizeBytes = null)
{
    public RegisterCustomerAttachmentCommand ToCommand(string customerId)
        => new()
        {
            CustomerId = customerId,
            Title = Title,
            OriginalFileName = OriginalFileName,
            StorageBucket = StorageBucket,
            StorageObjectKey = StorageObjectKey,
            ContentType = ContentType,
            SizeBytes = SizeBytes
        };
}
