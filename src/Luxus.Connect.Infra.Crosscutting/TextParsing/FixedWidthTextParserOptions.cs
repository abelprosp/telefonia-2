namespace Luxus.Connect.Infra.Crosscutting.TextParsing;

/// <summary>Opções para leitura do tipo de registro e fatias de campo.</summary>
public sealed class FixedWidthTextParserOptions
{
    /// <summary>Índice base-0 do código de registro na linha (ex.: Vivo TXT = 110).</summary>
    public int RecordTypeOffset { get; set; } = 110;

    /// <summary>Largura do código de registro (ex.: 4 → "010D").</summary>
    public int RecordTypeLength { get; set; } = 4;

    /// <summary>Aplicar <see cref="string.Trim"/> no valor de <c>RecordType</c> lido da linha.</summary>
    public bool TrimRecordType { get; set; } = true;

    /// <summary>Trim padrão em cada fatia de campo textual antes de converter.</summary>
    public bool TrimFieldValues { get; set; } = true;
}
