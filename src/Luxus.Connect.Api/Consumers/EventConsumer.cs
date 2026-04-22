using System.Diagnostics;
using Goal.Application.Events;
using MassTransit;
using MassTransit.Metadata;

namespace Luxus.Connect.Api.Consumers;

public abstract class EventConsumer<TEvent>(ILogger logger) : IConsumer<TEvent>
    where TEvent : class, IEvent
{
    protected virtual string ConsumerName { get; } = TypeMetadataCache<TEvent>.ShortName;

    public async Task Consume(ConsumeContext<TEvent> context)
    {
        var timer = Stopwatch.StartNew();

        try
        {
            await HandleEvent(context.Message, context.CancellationToken);

            timer.Stop();
            await context.NotifyConsumed(timer.Elapsed, ConsumerName);
        }
        catch (Exception ex)
        {
            timer.Stop();
            logger.LogError(ex, "{InformationData}: An error occurred while consuming an event.", ConsumerName);
            await context.NotifyFaulted(timer.Elapsed, ConsumerName, ex);
        }
    }

    protected abstract Task HandleEvent(TEvent @event, CancellationToken cancellationToken = default);
}