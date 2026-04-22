using System.Text.Json;
using Luxus.Connect.Infra.Crosscutting.Extensions;

namespace Luxus.Connect.Infra.Http.JsonNamePolicies;

public class JsonSnakeCaseNamingPolicy : JsonNamingPolicy
{
    public override string ConvertName(string name)
        => name.ToSnakeCase();
}
