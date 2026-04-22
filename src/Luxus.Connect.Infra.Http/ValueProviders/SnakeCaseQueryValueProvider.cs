using System.Globalization;
using Luxus.Connect.Infra.Crosscutting.Extensions;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc.ModelBinding;

namespace Luxus.Connect.Infra.Http.ValueProviders;

public class SnakeCaseQueryValueProvider(
    BindingSource bindingSource,
    IQueryCollection values,
    CultureInfo culture) : QueryStringValueProvider(bindingSource, values, culture), IValueProvider
{

    public override bool ContainsPrefix(string prefix)
        => base.ContainsPrefix(prefix.ToSnakeCase());

    public override ValueProviderResult GetValue(string key)
        => base.GetValue(key.ToSnakeCase());
}
