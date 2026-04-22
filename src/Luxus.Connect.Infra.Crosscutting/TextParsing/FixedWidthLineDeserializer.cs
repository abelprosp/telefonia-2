using System.Reflection;

namespace Luxus.Connect.Infra.Crosscutting.TextParsing;

internal interface IFixedWidthLineDeserializer
{
    Type ClrType { get; }
    object Deserialize(string line);
}

internal sealed class FixedWidthLineDeserializer<TRecord> : IFixedWidthLineDeserializer where TRecord : class, new()
{
    private readonly FixedWidthFieldBinding<TRecord>[] _bindings;
    private readonly FixedWidthTextParserOptions _options;
    private readonly PropertyInfo? _recordTypeProperty;
    private readonly bool _hasRecordTypeBinding;

    public FixedWidthLineDeserializer(
        IReadOnlyList<FixedWidthFieldBinding<TRecord>> bindings,
        FixedWidthTextParserOptions options)
    {
        _bindings = bindings.ToArray();
        _options = options;
        _recordTypeProperty = typeof(TRecord).GetProperty("RecordType", BindingFlags.Public | BindingFlags.Instance);
        _hasRecordTypeBinding = _bindings.Any(b => b.Property.Name == "RecordType");
    }

    public Type ClrType => typeof(TRecord);

    public object Deserialize(string line)
    {
        var record = new TRecord();
        foreach (FixedWidthFieldBinding<TRecord> binding in _bindings)
            binding.Apply(line, record);

        if (!_hasRecordTypeBinding
            && _recordTypeProperty is not null
            && _recordTypeProperty.PropertyType == typeof(string))
        {
            string? current = _recordTypeProperty.GetValue(record) as string;
            if (string.IsNullOrWhiteSpace(current))
            {
                string rt = ReadRecordTypeFromLine(line);
                _recordTypeProperty.SetValue(record, rt);
            }
        }

        return record;
    }

    private string ReadRecordTypeFromLine(string line)
    {
        int o = _options.RecordTypeOffset;
        int len = _options.RecordTypeLength;
        if (o < 0 || len <= 0 || o >= line.Length)
            return string.Empty;

        int take = Math.Min(len, line.Length - o);
        string s = line.Substring(o, take);
        return _options.TrimRecordType ? s.Trim() : s;
    }
}
