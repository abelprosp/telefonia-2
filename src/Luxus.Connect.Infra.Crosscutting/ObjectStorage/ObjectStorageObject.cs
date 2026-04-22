using Amazon.S3.Model;

namespace Luxus.Connect.Infra.Crosscutting.ObjectStorage;

public sealed class ObjectStorageObject : IAsyncDisposable
{
    private readonly GetObjectResponse? _response;
    private readonly Stream? _stream;

    internal ObjectStorageObject(GetObjectResponse response)
    {
        _response = response;
    }

    private ObjectStorageObject(Stream stream)
    {
        _stream = stream;
    }

    /// <summary>Conteúdo em memória (ex.: testes ou stubs) sem resposta S3.</summary>
    public static ObjectStorageObject FromStream(Stream stream)
    {
        ArgumentNullException.ThrowIfNull(stream);
        return new ObjectStorageObject(stream);
    }

    public Stream Stream
        => _response?.ResponseStream ?? _stream!;

    public string? ContentType => _response is null
        ? null
        : string.IsNullOrWhiteSpace(_response.Headers.ContentType)
            ? null
            : _response.Headers.ContentType;

    public long ContentLength =>
        _response != null
            ? _response.ContentLength
            : _stream is MemoryStream ms
                ? ms.Length
                : 0;

    public ValueTask DisposeAsync()
    {
        if (_response is not null)
        {
            _response.Dispose();
        }
        else if (_stream is not null)
        {
            return _stream.DisposeAsync();
        }

        return ValueTask.CompletedTask;
    }
}
