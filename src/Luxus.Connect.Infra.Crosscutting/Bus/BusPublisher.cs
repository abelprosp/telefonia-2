using MassTransit;

namespace Luxus.Connect.Infra.Crosscutting.Bus;

public sealed class BusPublisher : IBusPublisher
{
    private readonly IPublishEndpoint _bus;

    public BusPublisher(IPublishEndpoint bus)
    {
        _bus = bus;
    }

    public async Task Publish<T>(T message, CancellationToken cancellationToken = default)
        where T : class
        => await _bus.Publish(message, cancellationToken);
}
