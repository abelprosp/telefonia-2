using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Http.Controllers;
using Microsoft.AspNetCore.Diagnostics;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

namespace Luxus.Connect.Infra.Http.Handlers;

public sealed class ConnectExceptionHandler(ILogger<ConnectExceptionHandler> logger, IOptions<JsonOptions> options) : IExceptionHandler
{
    private readonly ILogger<ConnectExceptionHandler> _logger = logger;

    public async ValueTask<bool> TryHandleAsync(HttpContext httpContext, Exception exception, CancellationToken cancellationToken)
    {
        _logger.LogError(exception, "An unexpected problem has occurred: {InformationData}", exception.Message);

        httpContext.Response.StatusCode = StatusCodes.Status500InternalServerError;

        await httpContext.Response.WriteAsJsonAsync(
            ApiResponse.Fail(Notifications.Shared.UnexpectedError(exception.Message)),
            options.Value.JsonSerializerOptions,
            cancellationToken);

        return true;
    }
}