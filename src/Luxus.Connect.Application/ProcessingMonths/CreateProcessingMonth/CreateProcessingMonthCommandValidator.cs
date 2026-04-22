using FluentValidation;
using Luxus.Connect.Contracts.ProcessingMonths.Commands;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Validations.Fluent;

namespace Luxus.Connect.Application.ProcessingMonths.CreateProcessingMonth;

internal sealed class CreateProcessingMonthCommandValidator : AbstractValidator<CreateProcessingMonthCommand>
{
    public CreateProcessingMonthCommandValidator()
    {
        RuleFor(x => x.OrganizationId)
            .NotEmpty()
            .WithNotification(Notifications.Shared.ORGANIZATION_ID_REQUIRED);

        RuleFor(x => x.ProviderId)
            .NotEmpty()
            .WithNotification(Notifications.ProcessingMonths.PROCESSING_MONTH_PROVIDER_REQUIRED);

        RuleFor(x => x.Year)
            .InclusiveBetween(2000, 2100)
            .WithNotification(Notifications.ProcessingMonths.PROCESSING_MONTH_YEAR_INVALID);

        RuleFor(x => x.Month)
            .InclusiveBetween(1, 12)
            .WithNotification(Notifications.ProcessingMonths.PROCESSING_MONTH_MONTH_INVALID);

        RuleFor(x => x.DisplayName)
            .NotEmpty()
            .WithNotification(Notifications.ProcessingMonths.PROCESSING_MONTH_DISPLAY_NAME_REQUIRED)
            .MaximumLength(128)
            .WithNotification(Notifications.ProcessingMonths.PROCESSING_MONTH_DISPLAY_NAME_MAX_LENGTH);
    }
}
