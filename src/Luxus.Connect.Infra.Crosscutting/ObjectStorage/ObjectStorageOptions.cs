namespace Luxus.Connect.Infra.Crosscutting.ObjectStorage;

public sealed class ObjectStorageOptions
{
    public const string SectionName = "ObjectStorage";

    public string ServiceUrl { get; set; } = string.Empty;

    public string AccessKeyId { get; set; } = string.Empty;

    public string SecretAccessKey { get; set; } = string.Empty;

    public bool ForcePathStyle { get; set; } = true;
}
