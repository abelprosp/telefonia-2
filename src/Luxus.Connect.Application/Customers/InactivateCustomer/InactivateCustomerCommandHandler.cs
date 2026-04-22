using ConduitR.Abstractions;
using Goal.Application.Commands;
using Luxus.Connect.Contracts.Customers.Commands;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Application.Customers.InactivateCustomer;

internal sealed class InactivateCustomerCommandHandler(
    IAppUnitOfWork uow,
    AppState appState)
    : ICommandHandler<InactivateCustomerCommand, OneOf<None, AppError>>
    , IRequestHandler<InactivateCustomerCommand, OneOf<None, AppError>>
{
    public async ValueTask<OneOf<None, AppError>> Handle(InactivateCustomerCommand command, CancellationToken cancellationToken)
    {
        if (appState.User is null)
        {
            return new BusinessRuleError(Notifications.Shared.DOMAIN_VIOLATION);
        }

        Customer? entity = await uow.Customers.GetByIdAsync(
            appState.Organization!.Id,
            command.Id,
            cancellationToken);

        if (entity is null)
        {
            return new ResourceNotFoundError(Notifications.Customers.CUSTOMER_NOT_FOUND);
        }

        entity.Inactivate();

        await uow.CommitAsync(cancellationToken);

        return default(None);
    }
}
