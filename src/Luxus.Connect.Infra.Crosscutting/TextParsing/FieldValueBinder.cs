using System.Globalization;
using System.Reflection;

namespace Luxus.Connect.Infra.Crosscutting.TextParsing;

internal static class FieldValueBinder
{
    internal static Action<TRecord, string> CreateSetter<TRecord>(PropertyInfo property, string? format, bool trimFieldValues)
    {
        Type targetType = property.PropertyType;
        Type underlying = Nullable.GetUnderlyingType(targetType) ?? targetType;

        return (record, raw) =>
        {
            string s = trimFieldValues ? raw.Trim() : raw;
            if (string.IsNullOrWhiteSpace(s))
            {
                if (Nullable.GetUnderlyingType(targetType) is not null)
                    property.SetValue(record, null);
                else if (targetType == typeof(string))
                    property.SetValue(record, string.Empty);

                return;
            }

            object? value = ConvertSlice(underlying, s, format);
            property.SetValue(record, value);
        };
    }

    private static object? ConvertSlice(Type type, string s, string? format)
    {
        if (type == typeof(string))
            return s;

        if (type == typeof(char) || type == typeof(char?))
            return s.Length > 0 ? s[0] : type == typeof(char?) ? null : '\0';

        if (type == typeof(int))
            return int.Parse(s, CultureInfo.InvariantCulture);

        if (type == typeof(long))
            return long.Parse(s, CultureInfo.InvariantCulture);

        if (type == typeof(short))
            return short.Parse(s, CultureInfo.InvariantCulture);

        if (type == typeof(decimal))
            return decimal.Parse(s, NumberStyles.Number, CultureInfo.InvariantCulture);

        if (type == typeof(double))
            return double.Parse(s, NumberStyles.Number, CultureInfo.InvariantCulture);

        if (type == typeof(float))
            return float.Parse(s, NumberStyles.Number, CultureInfo.InvariantCulture);

        if (type == typeof(bool))
            return bool.Parse(s);

        if (type == typeof(DateOnly))
        {
            if (string.IsNullOrWhiteSpace(format))
                return DateOnly.Parse(s, CultureInfo.InvariantCulture);
            return DateOnly.ParseExact(s, NormalizeDateFormat(format), CultureInfo.InvariantCulture);
        }

        if (type == typeof(DateTime))
        {
            if (string.IsNullOrWhiteSpace(format))
                return DateTime.Parse(s, CultureInfo.InvariantCulture, DateTimeStyles.None);
            return DateTime.ParseExact(s, NormalizeDateFormat(format), CultureInfo.InvariantCulture, DateTimeStyles.None);
        }

        if (type.IsEnum)
            return Enum.Parse(type, s, ignoreCase: true);

        throw new NotSupportedException($"Tipo de propriedade não suportado pelo parser fixo: {type.Name}");
    }

    /// <summary>Normaliza formatos informados como "DDmmyyyy" → padrões do .NET.</summary>
    private static string NormalizeDateFormat(string format)
    {
        return format.Trim() switch
        {
            "DDmmyyyy" or "ddMMyyyy" => "ddMMyyyy",
            "DDMMYYYY" => "ddMMyyyy",
            "yyyyMMdd" => "yyyyMMdd",
            "YYYYMMDD" => "yyyyMMdd",
            _ => format,
        };
    }
}
