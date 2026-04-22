namespace Luxus.Connect.Contracts.ObjectStorage.Models;

public sealed record PresignedUrlModel(
    string Url,
    string HttpMethod,
    DateTimeOffset ExpiresAtUtc);
