using System.Net;
using Amazon.S3;
using Amazon.S3.Model;

namespace Luxus.Connect.Infra.Crosscutting.ObjectStorage;

/// <summary>Usa o <see cref="IAmazonS3"/> já registrado no container (ex.: Cloudflare R2).</summary>
public sealed class R2ObjectStorageClient(IAmazonS3 amazonS3) : IObjectStorageClient
{
    public PresignedUrlResult CreatePresignedUploadUrl(
        string bucketName,
        string objectKey,
        TimeSpan expires,
        string? contentType = null)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(bucketName);
        ArgumentException.ThrowIfNullOrWhiteSpace(objectKey);

        DateTime expiresAtUtc = DateTime.UtcNow.Add(expires);
        var request = new GetPreSignedUrlRequest
        {
            BucketName = bucketName,
            Key = objectKey,
            Verb = HttpVerb.PUT,
            Expires = expiresAtUtc,
        };

        if (!string.IsNullOrWhiteSpace(contentType))
        {
            request.ContentType = contentType;
        }

        string url = amazonS3.GetPreSignedURL(request);
        return new PresignedUrlResult(url, "PUT", new DateTimeOffset(expiresAtUtc, TimeSpan.Zero));
    }

    public PresignedUrlResult CreatePresignedDownloadUrl(
        string bucketName,
        string objectKey,
        TimeSpan expires)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(bucketName);
        ArgumentException.ThrowIfNullOrWhiteSpace(objectKey);

        DateTime expiresAtUtc = DateTime.UtcNow.Add(expires);
        var request = new GetPreSignedUrlRequest
        {
            BucketName = bucketName,
            Key = objectKey,
            Verb = HttpVerb.GET,
            Expires = expiresAtUtc,
        };

        string url = amazonS3.GetPreSignedURL(request);
        return new PresignedUrlResult(url, "GET", new DateTimeOffset(expiresAtUtc, TimeSpan.Zero));
    }

    public async Task<ObjectStorageObject> GetObjectAsync(
        string bucketName,
        string objectKey,
        CancellationToken cancellationToken = default)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(bucketName);
        ArgumentException.ThrowIfNullOrWhiteSpace(objectKey);

        try
        {
            GetObjectResponse response = await amazonS3.GetObjectAsync(
                new GetObjectRequest
                {
                    BucketName = bucketName,
                    Key = objectKey,
                },
                cancellationToken)
                .ConfigureAwait(false);

            return new ObjectStorageObject(response);
        }
        catch (AmazonS3Exception ex) when (
            string.Equals(ex.ErrorCode, "NoSuchKey", StringComparison.Ordinal)
            || ex.StatusCode == HttpStatusCode.NotFound)
        {
            throw new ObjectStorageObjectNotFoundException(bucketName, objectKey, ex);
        }
    }
}
