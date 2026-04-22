using Keycloak.AuthServices.Authentication;
using Microsoft.AspNetCore.Authentication;
using Microsoft.AspNetCore.OpenApi;
using Microsoft.Net.Http.Headers;
using Microsoft.OpenApi;

namespace Luxus.Connect.Api.Infra.OpenApi;

internal sealed class BearerSecuritySchemeTransformer(
    IAuthenticationSchemeProvider authenticationSchemeProvider,
    IConfiguration configuration)
    : IOpenApiDocumentTransformer
{
    public async Task TransformAsync(OpenApiDocument document, OpenApiDocumentTransformerContext context, CancellationToken cancellationToken)
    {
        IEnumerable<AuthenticationScheme> authenticationSchemes = await authenticationSchemeProvider.GetAllSchemesAsync();

        if (!authenticationSchemes.Any(authScheme => authScheme.Name == "Bearer"))
        {
            return;
        }

        KeycloakAuthenticationOptions? keycloakOptions = configuration
            .GetSection(KeycloakAuthenticationOptions.Section)
            .Get<KeycloakAuthenticationOptions>();

        if (keycloakOptions?.KeycloakUrlRealm is null)
        {
            return;
        }

        // Only proceed if Bearer authentication is configured
        // Define the Bearer security scheme
        var bearerScheme = new OpenApiSecurityScheme
        {
            Type = SecuritySchemeType.OAuth2,
            Scheme = "Bearer",
            In = ParameterLocation.Header,
            Name = HeaderNames.Authorization,
            Flows = new OpenApiOAuthFlows
            {
                Password = new OpenApiOAuthFlow
                {
                    TokenUrl = new Uri($"{keycloakOptions.KeycloakUrlRealm}protocol/openid-connect/token"),
                    Scopes = new Dictionary<string, string>()
                }
            }
        };

        // Ensure components are initialized
        document.Components ??= new OpenApiComponents();

        // Add the scheme to the document components
        document.AddComponent("Bearer", bearerScheme);

        // Create a security requirement referencing the scheme
        var securityRequirement = new OpenApiSecurityRequirement
        {
            [new OpenApiSecuritySchemeReference("Bearer", document)] = []
        };

        IEnumerable<KeyValuePair<HttpMethod, OpenApiOperation>> operations = document.Paths.Values
            .Where(p => p.Operations is not null)
            .SelectMany(p => p.Operations!);

        // Apply the requirement to all operations
        foreach (KeyValuePair<HttpMethod, OpenApiOperation> operation in operations)
        {
            operation.Value.Security ??= [];
            operation.Value.Security.Add(securityRequirement);
        }
    }
}
