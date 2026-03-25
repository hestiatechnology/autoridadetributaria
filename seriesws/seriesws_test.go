package seriesws

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/hestiatechnology/autoridadetributaria/seriesws/types"
)

func TestRegistarSerieSerialization(t *testing.T) {
	req := types.RegistarSerieRequest{
		Serie:                "TESTETETETOA",
		TipoSerie:            types.TipoSerieNormal,
		ClasseDoc:            types.ClasseDocSI,
		TipoDoc:              types.TipoDocFT,
		NumInicialSeq:        1,
		DataInicioPrevUtiliz: types.NewXSDDate(time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)),
		NumCertSWFatur:       3085,
		MeioProcessamento:    types.MeioProcessamentoPI,
	}

	bodyBytes, err := xml.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	env := soapEnvelope{
		Header: soapHeader{},
		Body:   soapBody{Content: bodyBytes},
	}
	envBytes, err := xml.MarshalIndent(env, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	xmlStr := string(envBytes)
	t.Log(xmlStr)
}
