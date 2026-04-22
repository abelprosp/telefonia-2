namespace Luxus.Connect.Contracts.ObjectStorage.Inputs;

public sealed record CreatePresignedDownloadUrlInput(
    string BucketName,
    string ObjectKey,
    int? ExpiresInSeconds = null);
