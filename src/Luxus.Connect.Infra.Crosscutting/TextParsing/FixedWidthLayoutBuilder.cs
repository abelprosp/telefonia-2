using System.Linq.Expressions;

namespace Luxus.Connect.Infra.Crosscutting.TextParsing;

/// <summary>
/// Configura campos de largura fixa para <typeparamref name="TRecord"/>.
/// Finalize com <see cref="And"/> para voltar ao <see cref="FixedWidthTextParser"/> e registrar outro tipo.
/// </summary>
public sealed class FixedWidthLayoutBuilder<TRecord> where TRecord : class, new()
{
    private readonly FixedWidthTextParser _parser;
    private readonly string _recordType;
    private readonly List<FixedWidthFieldBinding<TRecord>> _bindings = [];

    internal FixedWidthLayoutBuilder(FixedWidthTextParser parser, string recordType)
    {
        _parser = parser;
        _recordType = recordType;
    }

    /// <summary>Mapeia uma fatia da linha para uma propriedade.</summary>
    /// <param name="property">Propriedade de destino (expression tree).</param>
    /// <param name="offset">Índice base-0 inicial na linha.</param>
    /// <param name="length">Comprimento da fatia.</param>
    /// <param name="format">Para datas: padrão .NET (<c>yyyyMMdd</c>, <c>ddMMyyyy</c>) ou alias <c>DDmmyyyy</c>.</param>
    public FixedWidthLayoutBuilder<TRecord> Field<TProperty>(
        Expression<Func<TRecord, TProperty>> property,
        int offset,
        int length,
        string? format = null)
    {
        _bindings.Add(FixedWidthFieldBinding<TRecord>.Create(property, offset, length, format, _parser.Options.TrimFieldValues));
        return this;
    }

    /// <summary>Registra este layout e retorna o parser para encadear outro <c>Parse&lt;T&gt;(...)</c>.</summary>
    public FixedWidthTextParser And()
    {
        var deserializer = new FixedWidthLineDeserializer<TRecord>(_bindings, _parser.Options);
        _parser.RegisterDeserializer(_recordType, deserializer);

        return _parser;
    }
}
