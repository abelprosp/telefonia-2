using FluentValidation;
using Luxus.Connect.Contracts.Providers.Commands;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Validations.Fluent;

namespace Luxus.Connect.Application.Providers.CreateProvider;

internal sealed class CreateProviderCommandValidator : AbstractValidator<CreateProviderCommand>
{
    public CreateProviderCommandValidator()
    {
        RuleFor(x => x.OrganizationId)
            .NotEmpty().WithNotification(Notifications.Shared.ORGANIZATION_ID_REQUIRED)
            .MaximumLength(100).WithNotification(Notifications.Shared.ORGANIZATION_ID_REQUIRED);

        RuleFor(x => x.Name)
            .NotEmpty().WithNotification(Notifications.Providers.PROVIDER_NAME_REQUIRED)
            .MaximumLength(100).WithNotification(Notifications.Providers.PROVIDER_NAME_MAX_LENGTH);

        RuleFor(x => x.Slug)
            .NotEmpty().WithNotification(Notifications.Providers.PROVIDER_SLUG_REQUIRED)
            .MaximumLength(50).WithNotification(Notifications.Providers.PROVIDER_SLUG_MAX_LENGTH);
    }
}
