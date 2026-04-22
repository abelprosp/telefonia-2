using FluentValidation;
using Luxus.Connect.Infra.Crosscutting.Notifications;
using Luxus.Connect.Infra.Crosscutting.Validations.Fluent.Validators;

namespace Luxus.Connect.Infra.Crosscutting.Validations.Fluent;

public static class ValidationsExtensions
{
    public static IRuleBuilderOptions<T, string> Cnpj<T>(this IRuleBuilder<T, string> ruleBuilder)
        => ruleBuilder.SetValidator(new CnpjValidator<T>());

    public static IRuleBuilderOptions<T, string> Cpf<T>(this IRuleBuilder<T, string> ruleBuilder)
        => ruleBuilder.SetValidator(new CpfValidator<T>());

    public static IRuleBuilderOptions<T, TProperty> WithNotification<T, TProperty>(this IRuleBuilderOptions<T, TProperty> rule, Notification notification)
    {
        ArgumentNullException.ThrowIfNull(notification, nameof(notification));

        return rule
            .WithMessage(notification.Message)
            .WithErrorCode(notification.Code);
    }
}
