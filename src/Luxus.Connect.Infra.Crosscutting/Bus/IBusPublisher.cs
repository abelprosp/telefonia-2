namespace Luxus.Connect.Infra.Crosscutting.Bus;

public interface IBusPublisher
{
    Task Publish<T>(T message, CancellationToken cancellationToken = default)
        where T : class;
}
