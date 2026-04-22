using Asp.Versioning;
using Luxus.Connect.Contracts.ObjectStorage.Inputs;
using Luxus.Connect.Contracts.ObjectStorage.Models;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.ObjectStorage;
using Luxus.Connect.Infra.Http.Controllers;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;

namespace Luxus.Connect.Api.Features.Storage;

[ApiController]
[ApiVersion("1")]
[Authorize]
[Route("v{version:apiVersion}/[controller]")]
public class PreSignedUrlsController(IObjectStorageClient objectStorage) : ConnectApiController
{
    private const int DefaultExpiresSeconds = 900;
    private const int MinExpiresSeconds = 60;
    private const int MaxExpiresSeconds = 604800;

    [HttpPost("upload")]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public ActionResult<PresignedUrlModel> PostPresignedUploadUrl([FromBody] CreatePresignedUploadUrlInput input)
    {
        ActionResult? validation = ValidateBucketAndKey(input.BucketName, input.ObjectKey);

        if (validation is not null)
        {
            return validation;
        }

        if (!TryResolveExpiration(input.ExpiresInSeconds, out int expiresSeconds, out ActionResult? expiresError))
        {
            return expiresError!;
        }

        PresignedUrlResult result = objectStorage.CreatePresignedUploadUrl(
            input.BucketName.Trim(),
            input.ObjectKey.Trim(),
            TimeSpan.FromSeconds(expiresSeconds),
            input.ContentType);

        return Ok(ToModel(result));
    }

    [HttpPost("download")]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public ActionResult<PresignedUrlModel> PostPresignedDownloadUrl([FromBody] CreatePresignedDownloadUrlInput input)
    {
        ActionResult? validation = ValidateBucketAndKey(input.BucketName, input.ObjectKey);

        if (validation is not null)
        {
            return validation;
        }

        if (!TryResolveExpiration(input.ExpiresInSeconds, out int expiresSeconds, out ActionResult? expiresError))
        {
            return expiresError!;
        }

        PresignedUrlResult result = objectStorage.CreatePresignedDownloadUrl(
            input.BucketName.Trim(),
            input.ObjectKey.Trim(),
            TimeSpan.FromSeconds(expiresSeconds));

        return Ok(ToModel(result));
    }

    private static PresignedUrlModel ToModel(PresignedUrlResult result)
        => new(result.Url, result.HttpMethod, result.ExpiresAtUtc);

    private bool TryResolveExpiration(
        int? expiresInSeconds,
        out int seconds,
        out ActionResult? error)
    {
        seconds = expiresInSeconds ?? DefaultExpiresSeconds;

        if (seconds is <MinExpiresSeconds or >MaxExpiresSeconds)
        {
            error = BadRequest(ApiResponse.Fail(Notifications.ObjectStorage.PRESIGNED_EXPIRES_IN_SECONDS_INVALID));
            return false;
        }

        error = null;
        return true;
    }

    private ActionResult? ValidateBucketAndKey(string bucketName, string objectKey)
    {
        if (string.IsNullOrWhiteSpace(bucketName))
        {
            return BadRequest(ApiResponse.Fail(Notifications.InvoiceImports.STORAGE_BUCKET_REQUIRED));
        }

        if (bucketName.Length > 256)
        {
            return BadRequest(ApiResponse.Fail(Notifications.InvoiceImports.STORAGE_BUCKET_MAX_LENGTH));
        }

        if (string.IsNullOrWhiteSpace(objectKey))
        {
            return BadRequest(ApiResponse.Fail(Notifications.InvoiceImports.STORAGE_OBJECT_KEY_REQUIRED));
        }

        if (objectKey.Length > 2048)
        {
            return BadRequest(ApiResponse.Fail(Notifications.InvoiceImports.STORAGE_OBJECT_KEY_MAX_LENGTH));
        }

        if (IsInvalidObjectKey(objectKey))
        {
            return BadRequest(ApiResponse.Fail(Notifications.ObjectStorage.OBJECT_KEY_INVALID));
        }

        return null;
    }

    private static bool IsInvalidObjectKey(string objectKey)
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
