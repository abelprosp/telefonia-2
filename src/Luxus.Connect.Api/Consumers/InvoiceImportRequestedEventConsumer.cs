using Goal.Application.Commands;
using Luxus.Connect.Contracts.Providers.Commands;
using Luxus.Connect.Contracts.Providers.Events;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Crosscutting.Extensions;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Api.Consumers;

public sealed class InvoiceImportRequestedEventConsumer(
    ILogger<InvoiceImportRequestedEventConsumer> logger,
    IServiceScopeFactory scopeFactory)
    : EventConsumer<InvoiceImportRequestedEvent>(logger)
{
    protected override async Task HandleEvent(InvoiceImportRequestedEvent @event, CancellationToken cancellationToken)
    {
        await using AsyncServiceScope scope = scopeFactory.CreateAsyncScope();

        ICommandSender commandSender = scope.ServiceProvider.GetRequiredService<ICommandSender>();

        OneOf<None, AppError> result = await commandSender.Send(new ImportInvoiceCommand(@event.AggregateId), cancellationToken);

        if (result.IsError())
        {
            AppError error = result.GetError();

            logger.LogError(
                "Failed to process invoice import '{ImportRequestId}': {ErrorMessage}",
                @event.AggregateId,
                error);

            return;
        }

        logger.LogInformation(
            "Succesfully process invoice import '{ImportRequestId}'.",
            @event.AggregateId);
    }
}
