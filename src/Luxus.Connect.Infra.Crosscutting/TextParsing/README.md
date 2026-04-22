# Parser de texto em largura fixa (`FixedWidthTextParser`)

Registre cada `RecordType` com **generics**, **expressões** (`Expression<Func<>>`) e **API fluente**.

## Requisitos

- Propriedade **`RecordType`** (`string`) no modelo (para o predicado `r => r.RecordType == "010D"` e para preenchimento automático se não houver `.Field` para ela).
- Propriedades com **setter** público (ou `init`) para cada `.Field`.

## Exemplo

```csharp
using Luxus.Connect.Infra.Crosscutting.TextParsing;

var parser = new FixedWidthTextParser
{
    Options =
    {
        RecordTypeOffset = 110,
        RecordTypeLength = 4,
        TrimRecordType = true,
        TrimFieldValues = true,
    },
}
    .Parse<InvoiceHeader>(r => r.RecordType == "010D")
        .Field(r => r.AccountNumber, offset: 0, length: 10)
        .Field(r => r.DueDate, offset: 247, length: 8, format: "yyyyMMdd")
    .And()
    .Parse<InvoiceCustomer>("011D")
        .Field(r => r.Name, offset: 175, length: 50)
    .And();

object row = parser.DeserializeLine(line);
```

## Formatos de data

- Padrões .NET: `yyyyMMdd`, `ddMMyyyy`, etc.
- Alias aceito: `DDmmyyyy` → interpretado como `ddMMyyyy`.

## Métodos úteis

- `ReadRecordType(string line)` — lê o código na posição configurada.
- `TryDeserializeLine` / `DeserializeLine` — uma linha.
- `DeserializeLines(string text)` — várias linhas (ignora linhas cujo `RecordType` não foi registrado).
