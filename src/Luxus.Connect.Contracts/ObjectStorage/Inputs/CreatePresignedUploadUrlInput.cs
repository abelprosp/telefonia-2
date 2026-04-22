namespace Luxus.Connect.Contracts.ObjectStorage.Inputs;

public sealed record CreatePresignedUploadUrlInput(
    string BucketName,
    string ObjectKey,
    string? ContentType = null,
    int? ExpiresInSeconds = null);
