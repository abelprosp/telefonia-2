using FluentValidation;
using Luxus.Connect.Contracts.ProcessingMonths.Commands;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Validations.Fluent;

namespace Luxus.Connect.Application.ProcessingMonths.CloseProcessingMonthInContingency;

internal sealed class CloseProcessingMonthInContingencyCommandValidator : AbstractValidator<CloseProcessingMonthInContingencyCommand>
{
    public CloseProcessingMonthInContingencyCommandValidator()
    {
        RuleFor(x => x.OrganizationId)
            .NotEmpty()
            .WithNotification(Notifications.Shared.ORGANIZATION_ID_REQUIRED);

        RuleFor(x => x.Id)
            .NotEmpty()
            .WithNotification(Notifications.ProcessingMonths.PROCESSING_MONTH_ID_REQUIRED);

        RuleFor(x => x.Justification)
            .NotEmpty()
            .WithNotification(Notifications.ProcessingMonths.PROCESSING_MONTH_CONTINGENCY_JUSTIFICATION_REQUIRED)
            .MinimumLength(10)
            .WithNotification(Notifications.ProcessingMonths.PROCESSING_MONTH_CONTINGENCY_JUSTIFICATION_REQUIRED)
            .MaximumLength(4000)
            .WithNotification(Notifications.ProcessingMonths.PROCESSING_MONTH_CONTINGENCY_JUSTIFICATION_MAX_LENGTH);
    }
}
