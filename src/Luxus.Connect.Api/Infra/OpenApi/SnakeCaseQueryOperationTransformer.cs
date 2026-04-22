using Luxus.Connect.Infra.Crosscutting.Extensions;
using Microsoft.AspNetCore.OpenApi;
using Microsoft.OpenApi;

namespace Luxus.Connect.Api.Infra.OpenApi;

internal class SnakeCaseQueryOperationTransformer : IOpenApiOperationTransformer
{
    public Task TransformAsync(OpenApiOperation operation, OpenApiOperationTransformerContext context, CancellationToken cancellationToken)
    {
        if (operation.Parameters is null)
        {
            return Task.CompletedTask;
        }

        foreach (IOpenApiParameter parameter in operation.Parameters)
        {
            if (parameter.In == ParameterLocation.Query && parameter is OpenApiParameter queryParameter)
            {
                queryParameter.Name = queryParameter.Name!.ToSnakeCase();
            }
        }

        return Task.CompletedTask;
    }
}
