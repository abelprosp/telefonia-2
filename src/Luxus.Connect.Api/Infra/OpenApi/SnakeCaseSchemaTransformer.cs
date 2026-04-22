using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Infra.Crosscutting.Extensions;
using Microsoft.AspNetCore.OpenApi;
using Microsoft.OpenApi;

namespace Luxus.Connect.Api.Infra.OpenApi;

internal sealed class SnakeCaseSchemaTransformer : IOpenApiSchemaTransformer
{
    public Task TransformAsync(OpenApiSchema schema, OpenApiSchemaTransformerContext context, CancellationToken cancellationToken)
    {
        if (schema.Properties?.Count > 0)
        {
            schema.Properties = TransformProperties(schema.Properties);
        }

        return Task.CompletedTask;
    }

    private static Dictionary<string, IOpenApiSchema> TransformProperties(IDictionary<string, IOpenApiSchema> properties)
    {
        return properties?.ToDictionary(
            item => item.Key.ToSnakeCase(),
            item =>
            {
                if (item.Value.Properties?.Count > 0 && item.Value is OpenApiSchema value)
                {
                    value.Properties = TransformProperties(item.Value.Properties);
                }

                return item.Value;
            }
        ) ?? [];
    }
}
