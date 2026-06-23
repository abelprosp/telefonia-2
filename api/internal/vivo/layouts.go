package vivo

// DataType espelha o enum DataType de Layouts.cs.
type DataType int

const (
	DataTypeText DataType = iota
	DataTypeNumber
	DataTypeMoney
	DataTypeDate
	DataTypeTime
	DataTypeDatetime
)

// LayoutField descreve um campo de layout de largura fixa.
type LayoutField struct {
	Name   string
	Offset int
	Length int
	Type   DataType
	Format string
}

// Layout agrupa registros por tipo.
type Layout struct {
	Name      string
	Registers map[string][]LayoutField
}

// VivoLayout contém definições de layout VIVO (espelha Layouts.Vivo no C#).
// Nota: o layout 010D no C# está incompleto; o parser real usa VivoTextInvoiceParser.
var VivoLayout = map[string][]LayoutField{
	"010D": {
		{Name: "AccountNumber", Offset: 0, Length: 10, Type: DataTypeText},
		{Name: "EquipmentNumber", Offset: 10, Length: 15, Type: DataTypeText},
		{Name: "PhoneNumber", Offset: 25, Length: 16, Type: DataTypeText},
		{Name: "BlockNumber", Offset: 41, Length: 22, Type: DataTypeText},
		{Name: "Identifier", Offset: 25, Length: 8, Type: DataTypeDate, Format: "yyyyMMdd"},
		{Name: "BlockCode", Offset: 25, Length: 8, Type: DataTypeDate, Format: "yyyyMMdd"},
		{Name: "RecordType", Offset: 25, Length: 8, Type: DataTypeDate, Format: "yyyyMMdd"},
		{Name: "Qualifier", Offset: 25, Length: 8, Type: DataTypeDate, Format: "yyyyMMdd"},
	},
}
