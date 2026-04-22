using System.Diagnostics.CodeAnalysis;
using System.Linq.Expressions;

namespace Luxus.Connect.Infra.Crosscutting.TextParsing;

public sealed class FixedWidthTextParser
{
#pragma warning disable IDE0028 // Simplify collection initialization
    private readonly Dictionary<string, IFixedWidthLineDeserializer> _deserializers = new(StringComparer.Ordinal);
#pragma warning restore IDE0028 // Simplify collection initialization

    public FixedWidthTextParserOptions Options { get; } = new();

    /// <summary>Registra um tipo de registro a partir de <c>r => r.RecordType == "010D"</c>.</summary>
    public FixedWidthLayoutBuilder<TRecord> Parse<TRecord>(Expression<Func<TRecord, bool>> recordTypePredicate)
        where TRecord : class, new()
    {
        if (!RecordTypeExpression.TryGetExpectedRecordType(recordTypePredicate, out string recordType))
            throw new ArgumentException(
                "Informe uma expressão do tipo r => r.RecordType == \"XXXX\".",
                nameof(recordTypePredicate));

        return new FixedWidthLayoutBuilder<TRecord>(this, recordType);
    }

    /// <summary>Registra um tipo de registro pela chave literal (mesmo valor lido da linha).</summary>
    public FixedWidthLayoutBuilder<TRecord> Parse<TRecord>(string recordType) where TRecord : class, new()
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(recordType);
        return new FixedWidthLayoutBuilder<TRecord>(this, recordType.Trim());
    }

    internal void RegisterDeserializer(string recordType, IFixedWidthLineDeserializer deserializer)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(recordType);
        _deserializers[recordType] = deserializer;
    }

    /// <summary>Lê o valor de RecordType na linha conforme <see cref="FixedWidthTextParserOptions"/>.</summary>
    public string ReadRecordType(string line)
    {
        int o = Options.RecordTypeOffset;
        int len = Options.RecordTypeLength;

        if (string.IsNullOrWhiteSpace(line) || o < 0 || len <= 0 || o >= line.Length)
            return string.Empty;

        int take = Math.Min(len, line.Length - o);
        string s = line.Substring(o, take);

        return Options.TrimRecordType ? s.Trim() : s;
    }

    /// <summary>Deserializa uma linha; falha se o tipo de registro não estiver registrado.</summary>
    public object DeserializeLine(string line)
    {
        if (!TryDeserializeLine(line, out object? record) || record is null)
            throw new InvalidOperationException($"Tipo de registro não registrado: '{ReadRecordType(line)}'.");

        return record;
    }

    /// <summary>Tenta deserializar uma linha.</summary>
    public bool TryDeserializeLine(string line, [NotNullWhen(true)] out object? record)
    {
        record = null;

        if (string.IsNullOrWhiteSpace(line))
            return false;

        string rt = ReadRecordType(line);

        if (string.IsNullOrWhiteSpace(rt) || !_deserializers.TryGetValue(rt, out IFixedWidthLineDeserializer? d))
            return false;

        record = d.Deserialize(line);

        return true;
    }

    /// <summary>Deserializa todas as linhas não vazias do texto (quebras \r\n ou \n).</summary>
    public IEnumerable<object> DeserializeLines(string text)
    {
        if (string.IsNullOrWhiteSpace(text))
            yield break;

        foreach (string raw in text.Split(['\r', '\n'], StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries))
        {
            if (TryDeserializeLine(raw, out object? line))
                yield return line;
        }
    }

    /// <summary>Deserializa tentando converter para <typeparamref name="TRecord"/>.</summary>
    public bool TryDeserializeLine<TRecord>(string line, [NotNullWhen(true)] out TRecord? record) where TRecord : class, new()
    {
        record = null;
        if (!TryDeserializeLine(line, out object? obj) || obj is not TRecord typed)
            return false;

        record = typed;
        return true;
    }

    /// <summary>Tipos de registro já registrados (chaves).</summary>
    public IReadOnlyCollection<string> RegisteredRecordTypes => _deserializers.Keys;
}
