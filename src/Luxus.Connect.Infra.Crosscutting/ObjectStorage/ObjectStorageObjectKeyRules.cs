namespace Luxus.Connect.Infra.Crosscutting.ObjectStorage;

public static class ObjectStorageObjectKeyRules
{
    public static bool IsInvalid(string objectKey)
    {
        if (objectKey.Contains('\0', StringComparison.Ordinal))
        {
            return true;
        }

        if (objectKey.Contains("..", StringComparison.Ordinal))
        {
            return true;
        }

        return false;
    }
}
