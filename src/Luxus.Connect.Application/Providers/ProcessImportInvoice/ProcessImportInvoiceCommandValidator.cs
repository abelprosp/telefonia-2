using FluentValidation;
using Luxus.Connect.Contracts.Providers.Commands;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Validations.Fluent;

namespace Luxus.Connect.Application.Providers.ProcessImportInvoice;

internal sealed class ProcessImportInvoiceCommandValidator : AbstractValidator<ImportInvoiceCommand>
{
    public ProcessImportInvoiceCommandValidator()
    {
        RuleFor(x => x.ImportRequestId)
            .NotEmpty()
            .WithNotification(Notifications.InvoiceImports.IMPORT_REQUEST_ID_REQUIRED);
    }
}
