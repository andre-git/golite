package lib

import (
	"golite/internal/vdbe"
	"testing"
)

func TestDateFunctions(t *testing.T) {
	t.Run("date", func(t *testing.T) {
		res, _ := Date([]vdbe.Value{&stringValue{val: "2023-10-27"}})
		if res.String() != "2023-10-27" {
			t.Errorf("expected 2023-10-27, got %s", res.String())
		}
	})

	t.Run("datetime", func(t *testing.T) {
		res, _ := Datetime([]vdbe.Value{&stringValue{val: "2023-10-27 12:34:56"}})
		if res.String() != "2023-10-27 12:34:56" {
			t.Errorf("expected 2023-10-27 12:34:56, got %s", res.String())
		}
	})

	t.Run("modifiers", func(t *testing.T) {
		res, _ := Date([]vdbe.Value{
			&stringValue{val: "2023-10-27"},
			&stringValue{val: "+1 days"},
		})
		if res.String() != "2023-10-28" {
			t.Errorf("expected 2023-10-28, got %s", res.String())
		}
	})
}
