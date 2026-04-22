using System.Text.RegularExpressions;

namespace Luxus.Connect.Infra.Crosscutting.Constants;

public static partial class RegularExpressions
{
    [GeneratedRegex("[^0-9]")]
    public static partial Regex OnlyDigits();
}