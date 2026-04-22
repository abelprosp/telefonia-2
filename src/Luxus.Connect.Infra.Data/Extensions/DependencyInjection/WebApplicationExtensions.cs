using Microsoft.AspNetCore.Builder;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;

namespace Luxus.Connect.Infra.Data.Extensions.DependencyInjection;

public static class WebApplicationExtensions
{
    public static WebApplication MigrateDatabase(this WebApplication app)
    {
        using (IServiceScope serviceScope = app.Services.GetRequiredService<IServiceScopeFactory>().CreateScope())
        {
            try
            {
                AppDbContext? context = serviceScope.ServiceProvider.GetService<AppDbContext>();
                context?.Database.Migrate();
            }
            catch { }
        }

        return app;
    }
}