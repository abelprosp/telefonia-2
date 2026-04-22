using Microsoft.AspNetCore.Hosting;
using Microsoft.Extensions.Configuration;
using Serilog;
using Serilog.Exceptions;
using Serilog.Sinks.SystemConsole.Themes;

namespace Luxus.Connect.Infra.Crosscutting.Extensions;

public static class LoggingExtensions
{
    public static void ConfigureLogging(
        this LoggerConfiguration logger,
        IConfiguration configuration,
        IWebHostEnvironment environment)
    {
        logger
            .Enrich.FromLogContext()
            .Enrich.WithExceptionDetails()
            .Enrich.WithCorrelationId();

        if (environment is not null)
        {
            logger.Enrich.WithEnvironmentName();
        }

        logger
            .WriteTo.Console(outputTemplate: "[{Timestamp:HH:mm:ss} {Level}] {SourceContext}{NewLine}{Message:lj}{NewLine}{Exception}{NewLine}", theme: AnsiConsoleTheme.Code)
            .WriteTo.Seq(configuration.GetConnectionString("Seq")!)
            .ReadFrom.Configuration(configuration);
    }
}
