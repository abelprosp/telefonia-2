using Microsoft.AspNetCore.OpenApi;
using Microsoft.OpenApi;

namespace Luxus.Connect.Api.Infra.OpenApi;

internal sealed class ServerHostTransformer(IHttpContextAccessor httpContextAccessor) : IOpenApiDocumentTransformer
{
    public Task TransformAsync(OpenApiDocument document, OpenApiDocumentTransformerContext context, CancellationToken cancellationToken)
    {
        HttpContext? httpContext = httpContextAccessor.HttpContext;

        if (httpContext == null)
        {
            return Task.CompletedTask;
        }

        string scheme = httpContext.Request.Scheme ?? "http";
        string host = httpContext.Request.Host.Value ?? "localhost";

        document.Servers =
        [
            new OpenApiServer { Url = $"{scheme}://{host}" }
        ];

        return Task.CompletedTask;
    }
}
