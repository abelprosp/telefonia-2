namespace Luxus.Connect.Infra.Crosscutting.ObjectStorage;

/// <summary>URL assinada (S3 SigV4) para upload ou download direto no R2/S3.</summary>
public sealed record PresignedUrlResult(
    string Url,
    string HttpMethod,
    DateTimeOffset ExpiresAtUtc);
