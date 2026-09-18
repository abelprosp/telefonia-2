package models

// CustomerRegistrationProfile contains optional registration details. Personal
// identity/family fields are accepted only for individual customers.
type CustomerRegistrationProfile struct {
	ServiceUnit      string   `json:"service_unit,omitempty"`
	UA               string   `json:"ua,omitempty"`
	ShowUA           bool     `json:"show_ua"`
	Account          string   `json:"account,omitempty"`
	ShowAccount      bool     `json:"show_account"`
	Birthplace       string   `json:"birthplace,omitempty"`
	Sex              string   `json:"sex,omitempty"`
	IdentityType     string   `json:"identity_type,omitempty"`
	IdentityIssuedOn string   `json:"identity_issued_on,omitempty"`
	IdentityIssuer   string   `json:"identity_issuer,omitempty"`
	IdentityState    string   `json:"identity_state,omitempty"`
	RG               string   `json:"rg,omitempty"`
	FatherName       string   `json:"father_name,omitempty"`
	MotherName       string   `json:"mother_name,omitempty"`
	ContactPhone     string   `json:"contact_phone,omitempty"`
	Phone            string   `json:"phone,omitempty"`
	Mobile           string   `json:"mobile,omitempty"`
	Emails           []string `json:"emails,omitempty"`
	ContactName      string   `json:"contact_name,omitempty"`
	MaritalStatus    string   `json:"marital_status,omitempty"`
	SpouseCPF        string   `json:"spouse_cpf,omitempty"`
	Profession       string   `json:"profession,omitempty"`
	Notes            string   `json:"notes,omitempty"`
	ShowNotes        bool     `json:"show_notes"`
}
