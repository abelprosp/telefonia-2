namespace Luxus.Connect.Contracts.Customers.Responses;

public sealed record CustomerAttachmentResponse(
    string Id,
    string? Title,
    string OriginalFileName,
    string StorageBucket,
    string StorageObjectKey,
    string? ContentType,
    long? SizeBytes,
    DateTimeOffset UploadedAtUtc);
