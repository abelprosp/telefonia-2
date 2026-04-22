using Goal.Infra.Data.Query;

namespace Luxus.Connect.Infra.Data.Query.Repositories;

public abstract class AppQueryRepository(AppDbContext context)
    : QueryRepository
{
    private bool disposed;

    protected AppDbContext context = context;

    protected override void Dispose(bool disposing)
    {
        if (!disposed)
        {
            if (disposing)
            {
                context.Dispose();
            }

            context = null!;

            disposed = true;
        }
    }
}
