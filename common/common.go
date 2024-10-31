package common

// CountryCodes is a list of ISO 3166-1 alpha-2 country codes.
// Obtained via the UN Statistics Division.
// https://unstats.un.org/unsd/methodology/m49/overview/
var CountryCodes = []string{
	"AD", "AE", "AF", "AG", "AI", "AL", "AM", "AO", "AQ", "AR",
	"AS", "AT", "AU", "AW", "AX", "AZ", "BA", "BB", "BD", "BE",
	"BF", "BG", "BH", "BI", "BJ", "BL", "BM", "BN", "BO", "BQ",
	"BR", "BS", "BT", "BV", "BW", "BY", "BZ", "CA", "CC", "CD",
	"CF", "CG", "CH", "CI", "CK", "CL", "CM", "CN", "CO", "CR",
	"CU", "CV", "CW", "CX", "CY", "CZ", "DE", "DJ", "DK", "DM",
	"DO", "DZ", "EC", "EE", "EG", "EH", "ER", "ES", "ET", "FI",
	"FJ", "FK", "FM", "FO", "FR", "GA", "GB", "GD", "GE", "GF",
	"GG", "GH", "GI", "GL", "GM", "GN", "GP", "GQ", "GR", "GS",
	"GT", "GU", "GW", "GY", "HK", "HM", "HN", "HR", "HT", "HU",
	"ID", "IE", "IL", "IM", "IN", "IO", "IQ", "IR", "IS", "IT",
	"JE", "JM", "JO", "JP", "KE", "KG", "KH", "KI", "KM", "KN",
	"KP", "KR", "KW", "KY", "KZ", "LA", "LB", "LC", "LI", "LK",
	"LR", "LS", "LT", "LU", "LV", "LY", "MA", "MC", "MD", "ME",
	"MF", "MG", "MH", "MK", "ML", "MM", "MN", "MO", "MP", "MQ",
	"MR", "MS", "MT", "MU", "MV", "MW", "MX", "MY", "MZ", "NA",
	"NC", "NE", "NF", "NG", "NI", "NL", "NO", "NP", "NR", "NU",
	"NZ", "OM", "PA", "PE", "PF", "PG", "PH", "PK", "PL", "PM",
	"PN", "PR", "PS", "PT", "PW", "PY", "QA", "RE", "RO", "RS",
	"RU", "RW", "SA", "SB", "SC", "SD", "SE", "SG", "SH", "SI",
	"SJ", "SK", "SL", "SM", "SN", "SO", "SR", "SS", "ST", "SV",
	"SX", "SY", "SZ", "TC", "TD", "TF", "TG", "TH", "TJ", "TK",
	"TL", "TM", "TN", "TO", "TR", "TT", "TV", "TZ", "UA", "UG",
	"UM", "US", "UY", "UZ", "VA", "VC", "VE", "VG", "VI", "VN",
	"VU", "WF", "WS", "YE", "YT", "ZA", "ZM", "ZW",
}

var VatExemptionCodes = []string{
	"M01", "M02", "M04", "M05", "M06", "M07", "M09", "M10",
	"M11", "M12", "M13", "M14", "M15", "M16", "M19", "M20",
	"M21", "M25", "M26", "M30", "M31", "M32", "M33", "M34",
	"M40", "M41", "M42", "M43",
}

func ValidateNIFPT(nif string) bool {
	if len(nif) != 9 {
		return false
	}

	if nif[0] != '1' && nif[0] != '2' && nif[0] != '3' &&
		nif[:2] != "45" && nif[0] != '5' && nif[0] != '6' &&
		nif[:2] != "70" && nif[:2] != "71" && nif[:2] != "72" &&
		nif[:2] != "74" && nif[:2] != "75" && nif[:2] != "77" &&
		nif[:2] != "78" && nif[:2] != "79" && nif[0] != '8' &&
		nif[:2] != "90" && nif[:2] != "91" && nif[:2] != "98" &&
		nif[:2] != "99" {
		return false
	}

	check1 := int(nif[0]-'0') * 9
	check2 := int(nif[1]-'0') * 8
	check3 := int(nif[2]-'0') * 7
	check4 := int(nif[3]-'0') * 6
	check5 := int(nif[4]-'0') * 5
	check6 := int(nif[5]-'0') * 4
	check7 := int(nif[6]-'0') * 3
	check8 := int(nif[7]-'0') * 2

	total := check1 + check2 + check3 + check4 + check5 + check6 + check7 + check8
	mod11 := total % 11
	var comp int
	if mod11 == 1 || mod11 == 0 {
		comp = 0
	} else {
		comp = 11 - mod11
	}

	lastDigit := int(nif[8] - '0')
	return lastDigit == comp
}
