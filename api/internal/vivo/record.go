package vivo

// LineRecord campos comuns a todos os registros VIVO.
type LineRecord struct {
	AccountNumber string
	BlockNumber   string
	BlockCode     string
	RecordType    string
}

func bindCommon(line string, trim bool) LineRecord {
	return LineRecord{
		AccountNumber: bindString(line, 0, 10, trim),
		BlockNumber:   bindString(line, 61, 2, trim),
		BlockCode:     bindString(line, 84, 3, trim),
		RecordType:    bindString(line, 110, 4, trim),
	}
}

// Line marca tipos de registro parseados (equivalente a LineRecord no C#).
type Line interface {
	GetLineRecord() LineRecord
}

func (r LineRecord) GetLineRecord() LineRecord { return r }
