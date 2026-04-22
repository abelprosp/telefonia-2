using ConduitR.Abstractions;
using Goal.Application.Commands;
using Luxus.Connect.Contracts.Providers.Commands;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Application.Providers.InactivateProvider;

internal sealed class InactivateProviderCommandHandler(IAppUnitOfWork uow)
    : ICommandHandler<InactivateProviderCommand, OneOf<None, AppError>>
    , IRequestHandler<InactivateProviderCommand, OneOf<None, AppError>>
{
    public async ValueTask<OneOf<None, AppError>> Handle(InactivateProviderCommand command, CancellationToken cancellationToken)
    {
        Provider? entity = await uow.Providers.GetByIdAsync(
            command.OrganizationId,
            command.Id,
            cancellationToken);

        if (entity is null)
        {
            return new ResourceNotFoundError(Notifications.Providers.PROVIDER_NOT_FOUND);
        }

        entity.Inactivate();

        uow.Providers.Update(entity);
        await uow.CommitAsync(cancellationToken);

        return default(None);
    }
}
