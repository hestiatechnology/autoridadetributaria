package types

import (
	"encoding/xml"
	"time"
)

// XSDDate wraps time.Time and marshals/unmarshals as an xsd:date ("2006-01-02").
type XSDDate struct {
	time.Time
}

func NewXSDDate(t time.Time) XSDDate {
	return XSDDate{t.UTC().Truncate(24 * time.Hour)}
}

func (d XSDDate) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(d.Time.Format("2006-01-02"), start)
}

func (d *XSDDate) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := dec.DecodeElement(&s, &start); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

// SerieType - max 35 chars; identifies the document series.
type SerieType string

// TipoSerieType - 1 char.
// N = Normal, S = Autofaturação.
type TipoSerieType string

const (
	TipoSerieNormal TipoSerieType = "N"
	TipoSerieAuto   TipoSerieType = "S"
)

// ClasseDocType - 2 chars; document class.
// SI = Faturação, SS = Autofaturação, WD = Documentos de Trabalho,
// PY = Pagamentos, MG = Movimentos de Mercadorias.
type ClasseDocType string

const (
	ClasseDocSI ClasseDocType = "SI"
	ClasseDocSS ClasseDocType = "SS"
	ClasseDocWD ClasseDocType = "WD"
	ClasseDocPY ClasseDocType = "PY"
	ClasseDocMG ClasseDocType = "MG"
)

// TipoDocType - 2 chars; document type.
type TipoDocType string

const (
	// Faturação (SI/SS)
	TipoDocFT TipoDocType = "FT" // Fatura
	TipoDocFS TipoDocType = "FS" // Fatura Simplificada
	TipoDocFR TipoDocType = "FR" // Fatura-Recibo
	TipoDocNC TipoDocType = "NC" // Nota de Crédito
	TipoDocND TipoDocType = "ND" // Nota de Débito
	TipoDocRP TipoDocType = "RP" // Prémio ou Recibo de Prémio
	TipoDocRE TipoDocType = "RE" // Estorno ou Recibo de Estorno
	TipoDocCS TipoDocType = "CS" // Imputação a Co-seguradoras
	TipoDocLD TipoDocType = "LD" // Imputação a Co-seguradora Líder
	TipoDocRA TipoDocType = "RA" // Resseguro Aceite
	// Documentos de Trabalho (WD)
	TipoDocOR TipoDocType = "OR" // Orçamento
	TipoDocPF TipoDocType = "PF" // Proforma
	TipoDocCM TipoDocType = "CM" // Consultas de Mesa
	TipoDocCC TipoDocType = "CC" // Crédito de Consignação
	TipoDocFC TipoDocType = "FC" // Fatura de Consignação
	TipoDocFO TipoDocType = "FO" // Folhas de Obra
	TipoDocNE TipoDocType = "NE" // Nota de Encomenda
	TipoDocOU TipoDocType = "OU" // Outros
	// Pagamentos (PY)
	TipoDocRC TipoDocType = "RC" // Recibo (IVA de Caixa)
	// Movimentos de Mercadorias (MG)
	TipoDocGR TipoDocType = "GR" // Guia de Remessa
	TipoDocGT TipoDocType = "GT" // Guia de Transporte
	TipoDocGA TipoDocType = "GA" // Ativos Próprios
	TipoDocGC TipoDocType = "GC" // Consignação
	TipoDocGD TipoDocType = "GD" // Devoluções
)

// MeioProcessamentoType - 2 chars; document processing medium.
type MeioProcessamentoType string

const (
	MeioProcessamentoPI MeioProcessamentoType = "PI" // Programa de Faturação Integrado
	MeioProcessamentoPF MeioProcessamentoType = "PF" // Portal das Finanças
	MeioProcessamentoTC MeioProcessamentoType = "TC" // Máquina Registadora
	MeioProcessamentoOU MeioProcessamentoType = "OU" // Outro
)

// EstadoType - 1 char; series status.
type EstadoType string

const (
	EstadoAtivo      EstadoType = "A" // Ativo
	EstadoFinalizado EstadoType = "F" // Finalizado
	EstadoAnulado    EstadoType = "N" // Anulado
)

// MotivoType - 2 chars; reason for status change / cancellation.
type MotivoType string

const (
	MotivoER MotivoType = "ER" // Erro
)

// NifType - 9-digit Portuguese tax ID.
type NifType string

// CodValidacaoSerieType - exactly 8 chars; AT-assigned series validation code.
type CodValidacaoSerieType string

// -----------------------------------------------------------------------
// Request types
// -----------------------------------------------------------------------

// RegistarSerieRequest communicates a new document series to AT.
type RegistarSerieRequest struct {
	XMLName              xml.Name              `xml:"registarSerie"`
	ATNS                 string                `xml:"xmlns,attr"`
	Serie                SerieType             `xml:"serie"`
	TipoSerie            TipoSerieType         `xml:"tipoSerie"`
	ClasseDoc            ClasseDocType         `xml:"classeDoc"`
	TipoDoc              TipoDocType           `xml:"tipoDoc"`
	NumInicialSeq        int64                 `xml:"numInicialSeq"`
	DataInicioPrevUtiliz XSDDate               `xml:"dataInicioPrevUtiliz"`
	NumCertSWFatur       int                   `xml:"numCertSWFatur"`
	MeioProcessamento    MeioProcessamentoType `xml:"meioProcessamento"`
}

// FinalizarSerieRequest marks a series as finalized (no more documents will be issued).
type FinalizarSerieRequest struct {
	XMLName             xml.Name              `xml:"finalizarSerie"`
	ATNS                string                `xml:"xmlns,attr"`
	Serie               SerieType             `xml:"serie"`
	ClasseDoc           ClasseDocType         `xml:"classeDoc"`
	TipoDoc             TipoDocType           `xml:"tipoDoc"`
	CodValidacaoSerie   CodValidacaoSerieType `xml:"codValidacaoSerie"`
	SeqUltimoDocEmitido int64                 `xml:"seqUltimoDocEmitido"`
	Justificacao        string                `xml:"justificacao,omitempty"`
}

// ConsultarSeriesRequest queries registered series. All fields are optional filters.
type ConsultarSeriesRequest struct {
	XMLName           xml.Name               `xml:"consultarSeries"`
	ATNS              string                 `xml:"xmlns,attr"`
	Serie             *SerieType             `xml:"serie,omitempty"`
	TipoSerie         *TipoSerieType         `xml:"tipoSerie,omitempty"`
	ClasseDoc         *ClasseDocType         `xml:"classeDoc,omitempty"`
	TipoDoc           *TipoDocType           `xml:"tipoDoc,omitempty"`
	CodValidacaoSerie *CodValidacaoSerieType `xml:"codValidacaoSerie,omitempty"`
	DataRegistoDe     *XSDDate               `xml:"dataRegistoDe,omitempty"`
	DataRegistoAte    *XSDDate               `xml:"dataRegistoAte,omitempty"`
	Estado            *EstadoType            `xml:"estado,omitempty"`
	MeioProcessamento *MeioProcessamentoType `xml:"meioProcessamento,omitempty"`
}

// AnularSerieRequest cancels a previously registered series (correction of error).
type AnularSerieRequest struct {
	XMLName              xml.Name              `xml:"anularSerie"`
	ATNS                 string                `xml:"xmlns,attr"`
	Serie                SerieType             `xml:"serie"`
	ClasseDoc            ClasseDocType         `xml:"classeDoc"`
	TipoDoc              TipoDocType           `xml:"tipoDoc"`
	CodValidacaoSerie    CodValidacaoSerieType `xml:"codValidacaoSerie"`
	Motivo               MotivoType            `xml:"motivo"`
	// Must be true to confirm awareness that cancellation is only valid
	// if no documents have been issued with this series.
	DeclaracaoNaoEmissao bool `xml:"declaracaoNaoEmissao"`
}

// -----------------------------------------------------------------------
// Response types
// -----------------------------------------------------------------------

// OperationResultInfo contains the operation outcome code and message.
type OperationResultInfo struct {
	CodResultOper int    `xml:"codResultOper"`
	MsgResultOper string `xml:"msgResultOper"`
}

// SeriesInfo describes a registered series as returned by AT.
type SeriesInfo struct {
	Serie                SerieType             `xml:"serie"`
	TipoSerie            TipoSerieType         `xml:"tipoSerie"`
	ClasseDoc            ClasseDocType         `xml:"classeDoc"`
	TipoDoc              TipoDocType           `xml:"tipoDoc"`
	NumInicialSeq        int64                 `xml:"numInicialSeq"`
	NumFinalSeq          *int64                `xml:"numFinalSeq,omitempty"`
	DataInicioPrevUtiliz XSDDate               `xml:"dataInicioPrevUtiliz"`
	SeqUltimoDocEmitido  *int64                `xml:"seqUltimoDocEmitido,omitempty"`
	MeioProcessamento    MeioProcessamentoType `xml:"meioProcessamento"`
	NumCertSWFatur       int                   `xml:"numCertSWFatur"`
	CodValidacaoSerie    CodValidacaoSerieType `xml:"codValidacaoSerie"`
	DataRegisto          XSDDate               `xml:"dataRegisto"`
	Estado               EstadoType            `xml:"estado"`
	MotivoEstado         *MotivoType           `xml:"motivoEstado,omitempty"`
	Justificacao         *string               `xml:"justificacao,omitempty"`
	DataEstado           time.Time             `xml:"dataEstado"`
	NifComunicou         NifType               `xml:"nifComunicou"`
}

// SeriesResp is the common response payload for register/finalize/cancel operations.
type SeriesResp struct {
	InfoSerie      *SeriesInfo         `xml:"infoSerie,omitempty"`
	InfoResultOper OperationResultInfo `xml:"infoResultOper"`
}

// ConsultSeriesResp is the response payload for the query operation.
type ConsultSeriesResp struct {
	InfoSeries     []SeriesInfo        `xml:"infoSerie"`
	InfoResultOper OperationResultInfo `xml:"infoResultOper"`
}

// RegistarSerieResponse wraps the registarSerie operation response.
type RegistarSerieResponse struct {
	XMLName           xml.Name   `xml:"registarSerieResponse"`
	RegistarSerieResp SeriesResp `xml:"registarSerieResp"`
}

// FinalizarSerieResponse wraps the finalizarSerie operation response.
type FinalizarSerieResponse struct {
	XMLName            xml.Name   `xml:"finalizarSerieResponse"`
	FinalizarSerieResp SeriesResp `xml:"finalizarSerieResp"`
}

// ConsultarSeriesResponse wraps the consultarSeries operation response.
type ConsultarSeriesResponse struct {
	XMLName             xml.Name          `xml:"consultarSeriesResponse"`
	ConsultarSeriesResp ConsultSeriesResp `xml:"consultarSeriesResp"`
}

// AnularSerieResponse wraps the anularSerie operation response.
type AnularSerieResponse struct {
	XMLName         xml.Name   `xml:"anularSerieResponse"`
	AnularSerieResp SeriesResp `xml:"anularSerieResp"`
}
