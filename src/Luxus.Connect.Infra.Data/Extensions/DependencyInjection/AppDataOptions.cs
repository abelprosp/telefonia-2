namespace Luxus.Connect.Infra.Data.Extensions.DependencyInjection;

public sealed class AppDataOptions
{
    public string ConnectionString { get; private set; } = default!;

    public void UseConnectionString(string connectionString)
        => ConnectionString = connectionString;

}