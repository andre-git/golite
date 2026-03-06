package lib

import (
	"fmt"
	"golite/internal/vdbe"
	"strconv"
	"strings"
	"time"
)

const unixEpochJD = 2440587.5

func toJulianDay(t time.Time) float64 {
	return float64(t.UnixNano())/86400000000000.0 + unixEpochJD
}

func fromJulianDay(jd float64) time.Time {
	unixSeconds := (jd - unixEpochJD) * 86400.0
	sec := int64(unixSeconds)
	nsec := int64((unixSeconds - float64(sec)) * 1e9)
	return time.Unix(sec, nsec).UTC()
}

func parseDate(s string) (time.Time, error) {
	if s == "now" {
		return time.Now().UTC(), nil
	}
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		"15:04:05",
		"15:04",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	if jd, err := strconv.ParseFloat(s, 64); err == nil {
		return fromJulianDay(jd), nil
	}
	return time.Time{}, fmt.Errorf("invalid date format")
}

func applyModifiers(t time.Time, modifiers []vdbe.Value) (time.Time, error) {
	for _, mod := range modifiers {
		s := strings.ToLower(mod.String())
		parts := strings.Fields(s)
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "start":
			if len(parts) >= 3 && parts[1] == "of" {
				switch parts[2] {
				case "month":
					t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
				case "year":
					t = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
				case "day":
					t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
				}
			}
		case "utc":
			t = t.UTC()
		case "localtime":
			t = t.Local()
		default:
			if len(parts) >= 2 {
				val, err := strconv.ParseFloat(parts[0], 64)
				if err == nil {
					switch parts[1] {
					case "days":
						t = t.Add(time.Duration(val * 24 * float64(time.Hour)))
					case "hours":
						t = t.Add(time.Duration(val * float64(time.Hour)))
					case "minutes":
						t = t.Add(time.Duration(val * float64(time.Minute)))
					case "seconds":
						t = t.Add(time.Duration(val * float64(time.Second)))
					case "months":
						t = t.AddDate(0, int(val), 0)
					case "years":
						t = t.AddDate(int(val), 0, 0)
					}
				}
			}
		}
	}
	return t, nil
}

func Date(args []vdbe.Value) (vdbe.Value, error) {
	if len(args) == 0 { return nil, nil }
	t, err := parseDate(args[0].String())
	if err != nil { return nil, nil }
	t, _ = applyModifiers(t, args[1:])
	return &stringValue{t.Format("2006-01-02")}, nil
}

func Time(args []vdbe.Value) (vdbe.Value, error) {
	if len(args) == 0 { return nil, nil }
	t, err := parseDate(args[0].String())
	if err != nil { return nil, nil }
	t, _ = applyModifiers(t, args[1:])
	return &stringValue{t.Format("15:04:05")}, nil
}

func Datetime(args []vdbe.Value) (vdbe.Value, error) {
	if len(args) == 0 { return nil, nil }
	t, err := parseDate(args[0].String())
	if err != nil { return nil, nil }
	t, _ = applyModifiers(t, args[1:])
	return &stringValue{t.Format("2006-01-02 15:04:05")}, nil
}

func JulianDay(args []vdbe.Value) (vdbe.Value, error) {
	if len(args) == 0 { return nil, nil }
	t, err := parseDate(args[0].String())
	if err != nil { return nil, nil }
	t, _ = applyModifiers(t, args[1:])
	return &floatValue{toJulianDay(t)}, nil
}

func Strftime(args []vdbe.Value) (vdbe.Value, error) {
	if len(args) < 2 { return nil, nil }
	fmtStr := args[0].String()
	t, err := parseDate(args[1].String())
	if err != nil { return nil, nil }
	t, _ = applyModifiers(t, args[2:])
	res := fmtStr
	res = strings.ReplaceAll(res, "%Y", t.Format("2006"))
	res = strings.ReplaceAll(res, "%m", t.Format("01"))
	res = strings.ReplaceAll(res, "%d", t.Format("02"))
	res = strings.ReplaceAll(res, "%H", t.Format("15"))
	res = strings.ReplaceAll(res, "%M", t.Format("04"))
	res = strings.ReplaceAll(res, "%S", t.Format("05"))
	res = strings.ReplaceAll(res, "%J", fmt.Sprintf("%.5f", toJulianDay(t)))
	return &stringValue{res}, nil
}

type floatValue struct{ val float64 }
func (v *floatValue) Int64() int64 { return int64(v.val) }
func (v *floatValue) Float64() float64 { return v.val }
func (v *floatValue) String() string { return fmt.Sprintf("%g", v.val) }
func (v *floatValue) Blob() []byte { return nil }
func (v *floatValue) Type() byte { return 2 }
