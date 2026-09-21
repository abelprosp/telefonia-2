package axon

import "testing"

func TestGetLayout_unregistered(t *testing.T) {
	_, err := GetLayout("axon-v1")
	if err == nil {
		t.Fatal("expected error for unregistered layout")
	}
}

func TestRegisterAndValidate(t *testing.T) {
	err := RegisterLayout(Layout{
		Code:            "axon-sample",
		Description:     "sample from real file",
		RequiredHeaders: []string{"MSISDN", "Status"},
		Columns: map[string]string{
			"phone_number": "MSISDN",
			"status":       "Status",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	l, err := GetLayout("axon-sample")
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateHeaders(l, []string{"MSISDN", "Status", "Extra"}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateHeaders(l, []string{"Status"}); err == nil {
		t.Fatal("expected missing MSISDN")
	}
}
