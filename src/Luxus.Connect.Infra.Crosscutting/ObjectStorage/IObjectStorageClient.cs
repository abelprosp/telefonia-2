namespace Luxus.Connect.Infra.Crosscutting.ObjectStorage;

public interface IObjectStorageClient
{
    Task<ObjectStorageObject> GetObjectAsync(
        string bucketName,
        string objectKey,
        CancellationToken cancellationToken = default);

    /// <summary>Gera URL pré-assinada para <c>PUT</c> (upload do cliente).</summary>
    PresignedUrlResult CreatePresignedUploadUrl(
        string bucketName,
        string objectKey,
        TimeSpan expires,
        string? contentType = null);

    /// <summary>Gera URL pré-assinada para <c>GET</c> (download).</summary>
    PresignedUrlResult CreatePresignedDownloadUrl(
        string bucketName,
        string objectKey,
        TimeSpan expires);
}
