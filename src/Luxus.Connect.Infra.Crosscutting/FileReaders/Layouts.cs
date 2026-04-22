namespace Luxus.Connect.Infra.Crosscutting.FileReaders;

public static class Layouts
{
    public static Dictionary<string, List<LayoutField>> Vivo = new()
    {
        {
            "010D", new List<LayoutField>
            {
                new() { Name = "AccountNumber", Offset = 0, Length = 10, Type = DataType.Text },
                new() { Name = "EquipmentNumber", Offset = 10, Length = 15, Type = DataType.Text },
                new() { Name = "PhoneNumber", Offset = 25, Length = 16, Type = DataType.Text },
                new() { Name = "BlockNumber", Offset = 41, Length = 22, Type = DataType.Text },
                new() { Name = "Identifier", Offset = 25, Length = 8, Type = DataType.Date, Format = "yyyyMMdd" },
                new() { Name = "BlockCode", Offset = 25, Length = 8, Type = DataType.Date, Format = "yyyyMMdd" },
                new() { Name = "RecordType", Offset = 25, Length = 8, Type = DataType.Date, Format = "yyyyMMdd" },
                new() { Name = "Qualifier", Offset = 25, Length = 8, Type = DataType.Date, Format = "yyyyMMdd" }
            }
        }
    };
}

public class Layout
{
    public required string Name { get; set; }
    public required Dictionary<string, List<LayoutField>> Registers { get; set; }
}

public class LayoutField
{
    public required string Name { get; set; }
    public required int Offset { get; set; }
    public required int Length { get; set; }
    public required DataType Type { get; set; }
    public string? Format { get; set; }
}

public enum DataType
{
    Text,
    Number,
    Money,
    Date,
    Time,
    Datetime
}
