using System.Reflection;
using Amazon.Runtime;
using Amazon.S3;
using ConduitR;
using ConduitR.DependencyInjection;
using Goal.Application.Commands;
using Keycloak.AuthServices.Authentication;
using Keycloak.AuthServices.Authorization;
using Luxus.Connect.Application.Providers.CreateProvider;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.ObjectStorage;
using Luxus.Connect.Infra.Data.Extensions.DependencyInjection;
using Luxus.Connect.Infra.Data.Query.Extensions.DependencyInjection;
using Luxus.Connect.Infra.Http.Handlers;
using Luxus.Connect.Infra.IoC.Extensions.Options;
using MassTransit;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Options;

namespace Luxus.Connect.Infra.IoC.Extensions;

public static class ServiceColletionExtension
{
    public static IServiceCollection ConfigureServices(this IServiceCollection services, IConfiguration configuration)
    {
        services.AddHttpContextAccessor();
        services.AddScoped<AppState>();
        services.AddExceptionHandlers();

        services.AddCoreApplication(options =>
        {
            options.RegisterMediatorFromAssemblies(typeof(CreateProviderCommandHandler).Assembly);
        });

        services.AddAppData(options =>
        {
            options.UseConnectionString(configuration.GetConnectionString("DefaultConnection")!);
        });

        services.AddQueryData();
        services.AddObjectStorage(configuration);
        services.AddCoreDomain();

        return services;
    }

    public static IServiceCollection AddExceptionHandlers(this IServiceCollection services)
    {
        services.AddExceptionHandler<ConnectExceptionHandler>();
        services.AddProblemDetails();

        return services;
    }

    public static IServiceCollection AddKeycloak(this IServiceCollection services, IConfiguration configuration)
    {
        services
            .AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
            .AddKeycloakWebApi(configuration);

        services
            .AddAuthorization()
            .AddKeycloakAuthorization(options =>
            {
                options.EnableRolesMapping = RolesClaimTransformationSource.All;
                options.RolesResource = configuration["Keycloak:Resource"];
            })
            .AddAuthorizationBuilder()
            .AddPolicy("admin", policy =>
            {
                policy
                    .RequireAuthenticatedUser()
                    .RequireRole("admin");
            });

        return services;
    }

    public static IServiceCollection AddCoreApplication(this IServiceCollection services, Action<ApplicationOptions>? action = null)
    {
        var options = new ApplicationOptions();

        action?.Invoke(options);

        services.AddConduit(cfg =>
        {
            if (options.MediatorAssemblies.Length != 0)
            {
                cfg.AddHandlersFromAssemblies(options.MediatorAssemblies);
            }
            else
            {
                cfg.AddHandlersFromAssemblies(Assembly.GetExecutingAssembly());
            }

            cfg.PublishStrategy = PublishStrategy.Parallel; // Parallel (default), Sequential, StopOnFirstException
        });

        services.AddScoped<ICommandSender, CommandSender>();

        return services;
    }

    public static IServiceCollection AddCoreDomain(this IServiceCollection services, Action<DomainOptions>? action = null)
    {
        var options = new DomainOptions();

        action?.Invoke(options);

        return services;
    }

    public static IServiceCollection AddRabbitMq(this IServiceCollection services, Action<IBusRegistrationContext, IRabbitMqBusFactoryConfigurator>? configure = null)
    {
        services.AddMassTransit(x =>
        {
            x.AddConsumers(Assembly.GetEntryAssembly());

            x.UsingRabbitMq((context, configurator) =>
            {
                configure?.Invoke(context, configurator);
            });
        });

        return services;
    }

    public static IServiceCollection AddObjectStorageClient(this IServiceCollection services)
    {
        services.AddScoped<IObjectStorageClient, R2ObjectStorageClient>();
        return services;
    }

    public static IServiceCollection AddAmazonS3Client(this IServiceCollection services, IConfiguration configuration)
    {
        services.Configure<ObjectStorageOptions>(configuration.GetSection(ObjectStorageOptions.SectionName));

        services.AddScoped<IAmazonS3>(sp =>
        {
            ObjectStorageOptions opt = sp.GetRequiredService<IOptions<ObjectStorageOptions>>().Value;

            if (string.IsNullOrWhiteSpace(opt.ServiceUrl))
                throw new InvalidOperationException($"{ObjectStorageOptions.SectionName}:{nameof(ObjectStorageOptions.ServiceUrl)} is required.");

            if (string.IsNullOrWhiteSpace(opt.AccessKeyId) || string.IsNullOrWhiteSpace(opt.SecretAccessKey))
                throw new InvalidOperationException($"{ObjectStorageOptions.SectionName}: AccessKeyId and SecretAccessKey are required.");

            AWSCredentials credentials = new BasicAWSCredentials(opt.AccessKeyId, opt.SecretAccessKey);

            var config = new AmazonS3Config
            {
                ServiceURL = opt.ServiceUrl.TrimEnd('/'),
                ForcePathStyle = opt.ForcePathStyle,
            };

            return new AmazonS3Client(credentials, config);
        });

        return services;
    }

    public static IServiceCollection AddObjectStorage(this IServiceCollection services, IConfiguration configuration)
    {
        services
            .AddAmazonS3Client(configuration)
            .AddObjectStorageClient();

        return services;
    }
}
