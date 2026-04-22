using FluentValidation;
using FluentValidation.Validators;

namespace Luxus.Connect.Infra.Crosscutting.Validations.Fluent.Validators;

internal class CnpjValidator<T> : PropertyValidator<T, string>, IPropertyValidator
{
    public override bool IsValid(ValidationContext<T> context, string cnpj)
        => CustomValidations.IsValidCnpj(cnpj);

    public override string Name => "CnpjValidator";

    protected override string GetDefaultMessageTemplate(string errorCode)
        => "O Cnpj informado não é válido";
}
