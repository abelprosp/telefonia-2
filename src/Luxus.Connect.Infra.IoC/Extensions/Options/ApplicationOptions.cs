using System.Reflection;

namespace Luxus.Connect.Infra.IoC.Extensions.Options;

public sealed class ApplicationOptions
{
    public Assembly[] MediatorAssemblies { get; private set; } = [];

    public void RegisterMediatorFromAssemblies(params Assembly[] assemblies)
        => MediatorAssemblies = assemblies ?? [];
}
