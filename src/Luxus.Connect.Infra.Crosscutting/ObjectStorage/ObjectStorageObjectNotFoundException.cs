namespace Luxus.Connect.Infra.Crosscutting.ObjectStorage;

public sealed class ObjectStorageObjectNotFoundException(string bucketName, string objectKey, Exception? inner = null)
    : Exception($"Object not found in storage (bucket: {bucketName}, key: {objectKey}).", inner)
{
    public string BucketName { get; } = bucketName;

    public string ObjectKey { get; } = objectKey;
}
