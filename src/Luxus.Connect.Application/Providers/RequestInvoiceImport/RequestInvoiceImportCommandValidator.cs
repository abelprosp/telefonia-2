using FluentValidation;
using Luxus.Connect.Contracts.Providers.Commands;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Validations.Fluent;

namespace Luxus.Connect.Application.Providers.RequestInvoiceImport;

internal sealed class RequestInvoiceImportCommandValidator : AbstractValidator<ProviderInvoiceImportRequestCommand>
{
    public RequestInvoiceImportCommandValidator()
    {
        RuleFor(x => x.OrganizationId)
            .NotEmpty()
            .WithNotification(Notifications.Shared.ORGANIZATION_ID_REQUIRED);

        RuleFor(x => x.ProviderId)
            .NotEmpty()
            .WithNotification(Notifications.InvoiceImports.PROVIDER_ID_REQUIRED);

        RuleFor(x => x.ProcessingMonthId)
            .NotEmpty()
            .WithNotification(Notifications.InvoiceImports.PROCESSING_MONTH_ID_REQUIRED);

        RuleFor(x => x.StorageBucket)
            .NotEmpty()
            .WithNotification(Notifications.InvoiceImports.STORAGE_BUCKET_REQUIRED)
            .MaximumLength(256)
            .WithNotification(Notifications.InvoiceImports.STORAGE_BUCKET_MAX_LENGTH);

        RuleFor(x => x.StorageObjectKey)
            .NotEmpty()
            .WithNotification(Notifications.InvoiceImports.STORAGE_OBJECT_KEY_REQUIRED)
            .MaximumLength(2048)
            .WithNotification(Notifications.InvoiceImports.STORAGE_OBJECT_KEY_MAX_LENGTH);

        RuleFor(x => x.OriginalFileName)
            .MaximumLength(512)
            .When(x => x.OriginalFileName is not null)
            .WithNotification(Notifications.InvoiceImports.ORIGINAL_FILE_NAME_MAX_LENGTH);
    }
}
