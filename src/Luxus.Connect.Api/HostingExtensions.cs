using System.Text.Json;
using System.Text.Json.Serialization;
using Asp.Versioning;
using Goal.Infra.Crosscutting.Localization;
using Luxus.Connect.Api.Consumers;
using Luxus.Connect.Api.Infra.OpenApi;
using Luxus.Connect.Contracts.Providers.Events;
using Luxus.Connect.Infra.Crosscutting.Bus;
using Luxus.Connect.Infra.Crosscutting.Extensions;
using Luxus.Connect.Infra.Data.Extensions.DependencyInjection;
using Luxus.Connect.Infra.Http.JsonNamePolicies;
using Luxus.Connect.Infra.Http.ParameterTransformers;
using Luxus.Connect.Infra.Http.ValueProviders;
using Luxus.Connect.Infra.IoC.Extensions;
using MassTransit;
using Microsoft.AspNetCore.HttpOverrides;
using Microsoft.AspNetCore.Localization;
using Microsoft.AspNetCore.Mvc.ApplicationModels;
using Microsoft.IdentityModel.Logging;
using Microsoft.OpenApi;
using Scalar.AspNetCore;
using Serilog;

namespace Luxus.Connect.Api;

public static class HostingExtensions
{
    public static WebApplication ConfigureServices(this WebApplicationBuilder builder)
    {
        builder.Host.UseSerilog((_, lc) => lc.ConfigureLogging(builder.Configuration, builder.Environment));

        builder.Services.ConfigureServices(builder.Configuration);

        builder.Services.AddRabbitMq((context, configurator) =>
        {
            configurator.Host(builder.Configuration.GetConnectionString("RabbitMQ"));

            configurator.ReceiveEndpoint("luxus-connect-events", e =>
            {
                e.ConfigureConsumeTopology = false; // bind manual à exchange da mensagem (topic)

                e.Consumer<InvoiceImportRequestedEventConsumer>(context);

                e.Bind<InvoiceImportRequestedEvent>(x =>
                {
                    x.ExchangeType = RabbitMQ.Client.ExchangeType.Topic;
                    x.RoutingKey = $"{InvoiceImportRequestedEvent.RabbitMqRoutingKey}.#";
                });
            });

            configurator.Publish<InvoiceImportRequestedEvent>(cfg => cfg.ExchangeType = RabbitMQ.Client.ExchangeType.Topic);
            configurator.Send<InvoiceImportRequestedEvent>(cfg =>
                cfg.UseRoutingKeyFormatter(c =>
                    $"{InvoiceImportRequestedEvent.RabbitMqRoutingKey}.{c.Message.AggregateId}"));

            configurator.Publish<InvoiceImportMatrixAlertEvent>(cfg => cfg.ExchangeType = RabbitMQ.Client.ExchangeType.Topic);
            configurator.Send<InvoiceImportMatrixAlertEvent>(cfg =>
                cfg.UseRoutingKeyFormatter(c =>
                    $"{InvoiceImportMatrixAlertEvent.RabbitMqRoutingKey}.{c.Message.AggregateId}"));

            // Evita segundo endpoint automático para consumers já configurados acima.
            configurator.ConfigureEndpoints(context, f =>
            {
                f.Exclude<InvoiceImportRequestedEventConsumer>();
            });

            configurator.UseMessageRetry(retry => retry.Interval(3, TimeSpan.FromSeconds(5)));
        });

        builder.Services.AddScoped<IBusPublisher, BusPublisher>();

        builder.Services
            .AddApiVersioning(options =>
            {
                options.ReportApiVersions = true;
                options.AssumeDefaultVersionWhenUnspecified = true;
                options.DefaultApiVersion = new ApiVersion(1, 0);
            })
            .AddApiExplorer(options =>
            {
                options.GroupNameFormat = "'v'VVV";
                options.SubstituteApiVersionInUrl = true;
            });

        builder.Services
            .AddRouting(options => options.LowercaseUrls = true)
            .AddControllers(options =>
            {
                options.EnableEndpointRouting = false;
                options.ValueProviderFactories.Add(new SnakeCaseQueryValueProviderFactory());
                options.Conventions.Add(new RouteTokenTransformerConvention(new ToKebabParameterTransformer()));
            })
            .AddJsonOptions(options =>
            {
                options.JsonSerializerOptions.PropertyNamingPolicy = new JsonSnakeCaseNamingPolicy();
                options.JsonSerializerOptions.NumberHandling = JsonNumberHandling.Strict;
                options.JsonSerializerOptions.Converters.Add(new JsonStringEnumConverter());
            });

        builder.Services.ConfigureHttpJsonOptions(options =>
        {
            options.SerializerOptions.PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower;
        });

        builder.Services.AddEndpointsApiExplorer();

        builder.Services.AddOpenApi("v1", options =>
        {
            NumberTypeTransformer.MapType<decimal>(new OpenApiSchema { Type = JsonSchemaType.Number, Format = "decimal" });
            NumberTypeTransformer.MapType<decimal?>(new OpenApiSchema { Type = JsonSchemaType.Number | JsonSchemaType.Null, Format = "decimal" });
            NumberTypeTransformer.MapType<double>(new OpenApiSchema { Type = JsonSchemaType.Number, Format = "double" });
            NumberTypeTransformer.MapType<double?>(new OpenApiSchema { Type = JsonSchemaType.Number | JsonSchemaType.Null, Format = "double" });
            NumberTypeTransformer.MapType<int>(new OpenApiSchema { Type = JsonSchemaType.Integer, Format = "int32" });
            NumberTypeTransformer.MapType<int?>(new OpenApiSchema { Type = JsonSchemaType.Integer | JsonSchemaType.Null, Format = "int32" });
            NumberTypeTransformer.MapType<long>(new OpenApiSchema { Type = JsonSchemaType.Integer, Format = "int64" });
            NumberTypeTransformer.MapType<long?>(new OpenApiSchema { Type = JsonSchemaType.Integer | JsonSchemaType.Null, Format = "int64" });

            options.AddDocumentTransformer<BearerSecuritySchemeTransformer>();
            options.AddDocumentTransformer<ServerHostTransformer>();
            options.AddOperationTransformer<SnakeCaseQueryOperationTransformer>();
            options.AddSchemaTransformer<SnakeCaseSchemaTransformer>();
            options.AddSchemaTransformer<NumberTypeTransformer>();
        });

        builder.Services.AddKeycloak(builder.Configuration);

        builder.Services.Configure<ForwardedHeadersOptions>(options =>
        {
            options.ForwardedHeaders = ForwardedHeaders.XForwardedFor | ForwardedHeaders.XForwardedProto;
            options.KnownIPNetworks.Clear();
            options.KnownProxies.Clear();
        });

        builder.Services.AddCors(options =>
        {
            options.AddPolicy("Development", policyBuilder =>
            {
                policyBuilder
                    .AllowAnyOrigin()
                    .AllowAnyHeader()
                    .AllowAnyMethod();
            });

            options.AddPolicy("Staging", policyBuilder =>
            {
                string[] origins = (builder.Configuration["Cors:Origins"] ?? string.Empty)
                    .Split(';', StringSplitOptions.RemoveEmptyEntries);

                policyBuilder
                    .WithOrigins(origins)
                    .AllowAnyHeader()
                    .AllowAnyMethod()
                    .AllowCredentials();
            });

            options.AddPolicy("Production", policyBuilder =>
            {
                string[] origins = (builder.Configuration["Cors:Origins"] ?? string.Empty)
                    .Split(';', StringSplitOptions.RemoveEmptyEntries);

                policyBuilder
                    .WithOrigins(origins)
                    .AllowAnyHeader()
                    .AllowAnyMethod()
                    .AllowCredentials();
            });
        });

        return builder.Build();
    }

    public static WebApplication ConfigurePipeline(this WebApplication app)
    {
        app.UseForwardedHeaders();

        app.UseRequestLocalization(new RequestLocalizationOptions
        {
            DefaultRequestCulture = new RequestCulture(ApplicationCultures.English, ApplicationCultures.English),
            SupportedCultures =
            [
                ApplicationCultures.English
            ],
            SupportedUICultures =
            [
                ApplicationCultures.English
            ]
        });

        app.UseSerilogRequestLogging();
        app.UseExceptionHandler();

        if (app.Environment.IsDevelopment())
        {
            IdentityModelEventSource.ShowPII = true;

            app.MapOpenApi();
            app.MapScalarApiReference("api-docs", options =>
            {
                // Fluent API
                options
                    .WithTitle("Luxus Connect API")
                    .WithDefaultHttpClient(ScalarTarget.Http, ScalarClient.Http11)
                    .AddPreferredSecuritySchemes("Bearer")
                    .AddPasswordFlow("Bearer", flow =>
                    {
                        flow.TokenUrl = $"{app.Configuration["Keycloak:AuthServerUrl"]}/realms/{app.Configuration["Keycloak:Realm"]}/protocol/openid-connect/token";
                        flow.ClientId = app.Configuration["Keycloak:Resource"];
                        flow.SelectedScopes = app.Configuration["Keycloak:Scopes"]?.Split(' ');
                    });
            });
        }

        // Produção: sem UseHttpsRedirection — o TLS termina no reverse proxy (nginx).

        app.MapStaticAssets();
        app.UseRouting();

        app.UseAuthentication();

        app.UseCors(app.Environment.EnvironmentName);

        app.UseAuthorization();
        app.MapControllers();

        app.MigrateDatabase();

        return app;
    }
}