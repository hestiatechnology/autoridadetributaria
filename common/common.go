package common

import "errors"

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

var CountryCodesPTRegions = append(append(CountryCodes, "PT-AC"), "PT-MA")

type VatExemptionCode struct {
	Code        string
	Description string
	Law         string
}

var VatExemptionCodes = []VatExemptionCode{
	{Code: "M01", Description: "Artigo 16.º, n.º 6 do CIVA", Law: "Artigo 16.º, n.º 6, alíneas a) a d) do CIVA"},
	{Code: "M02", Description: "Artigo 6.º do Decreto-Lei n.º 198/90, de 19 dejunho", Law: "Artigo 6.º do Decreto‐Lei n.º 198/90, de 19 de junho"},
	{Code: "M04", Description: "Isento artigo 13.º do CIVA", Law: "Artigo 13.º do CIVA"},
	{Code: "M05", Description: "Isento artigo 14.º do CIVA", Law: "Artigo 14.º do CIVA"},
	{Code: "M06", Description: "Isento artigo 15.º do CIVA", Law: "Artigo 15.º do CIVA"},
	{Code: "M07", Description: "Isento artigo 9.º do CIVA", Law: "Artigo 9.º do CIVA"},
	{Code: "M09", Description: "IVA - não confere direito a dedução", Law: "Artigo 62.º alínea b) do CIVA"},
	{Code: "M10", Description: "IVA – regime de isenção", Law: "Artigo 57.º do CIVA"},
	{Code: "M11", Description: "Regime particular do tabaco", Law: "Decreto-Lei n.º 346/85, de 23 de agosto"},
	{Code: "M12", Description: "Regime da margem de lucro – Agências de viagens", Law: "Decreto-Lei n.º 221/85, de 3 de julho"},
	{Code: "M13", Description: "Regime da margem de lucro – Bens em segunda mão", Law: "Decreto-Lei n.º 199/96, de 18 de outubro"},
	{Code: "M14", Description: "Regime da margem de lucro – Objetos de arte", Law: "Decreto-Lei n.º 199/96, de 18 de outubro"},
	{Code: "M15", Description: "Regime da margem de lucro – Objetos de coleção e antiguidades", Law: "Decreto-Lei n.º 199/96, de 18 de outubro"},
	{Code: "M16", Description: "Isento artigo 14.º do RITI", Law: "Artigo 14.º do RITI"},
	{Code: "M19", Description: "Outras isenções", Law: "Isenções temporárias determinadas em diploma próprio"},
	{Code: "M20", Description: "IVA - regime forfetário", Law: "Artigo 59.º-D n.º2 do CIVA"},
	{Code: "M21", Description: "IVA – não confere direito à dedução (ou expressão similar)", Law: "Artigo 72.º n.º 4 do CIVA"},
	{Code: "M25", Description: "Mercadorias à consignação", Law: "Artigo 38.º n.º 1 alínea a) do CIVA"},
	{Code: "M26", Description: "Isenção de IVA com direito à dedução no cabaz alimentar", Law: "Lei n.º 17/2023, de 14 de abril"},
	{Code: "M30", Description: "IVA - autoliquidação", Law: "Artigo 2.º n.º 1 alínea i) do CIVA"},
	{Code: "M31", Description: "IVA - autoliquidação", Law: "Artigo 2.º n.º 1 alínea j) do CIVA"},
	{Code: "M32", Description: "IVA - autoliquidação", Law: "Artigo 2.º n.º 1 alínea l) do CIVA"},
	{Code: "M33", Description: "IVA - autoliquidação", Law: "Artigo 2.º n.º 1 alínea m) do CIVA"},
	{Code: "M34", Description: "IVA - autoliquidação", Law: "Artigo 2.º n.º 1 alínea n) do CIVA"},
	{Code: "M40", Description: "IVA - autoliquidação", Law: "Artigo 6.º n.º 6 alínea a) do CIVA, a contrário"},
	{Code: "M41", Description: "IVA - autoliquidação", Law: "Artigo 8.º n.º 3 do RITI"},
	{Code: "M42", Description: "IVA - autoliquidação", Law: "Decreto-Lei n.º 21/2007, de 29 de janeiro"},
	{Code: "M43", Description: "IVA - autoliquidação", Law: "Decreto-Lei n.º 362/99, de 16 de setembro"},
	{Code: "M99", Description: "Não sujeito ou não tributado", Law: "Outras situações de não liquidação do imposto (Exemplos: artigo 2.º, n.º 2 ; artigo 3.º, n.ºs 4, 6 e 7; artigo 4.º, n.º 5, todos do CIVA)"},
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

var (
	ErrInvalidNIFPT = errors.New("invalid nif")
)
