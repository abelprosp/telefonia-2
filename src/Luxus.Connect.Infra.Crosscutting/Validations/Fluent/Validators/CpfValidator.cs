using FluentValidation;
using FluentValidation.Validators;

namespace Luxus.Connect.Infra.Crosscutting.Validations.Fluent.Validators;

internal class CpfValidator<T> : PropertyValidator<T, string>, IPropertyValidator
{
    public override bool IsValid(ValidationContext<T> context, string cnpj)
        => CustomValidations.IsValidCpf(cnpj);

    public override string Name => "CpfValidator";

    protected override string GetDefaultMessageTemplate(string errorCode)
        => "O Cpf informado não é válido";
}
