package send

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type PollRequest struct {
	BaseRequest
	Question  string   `json:"question" form:"question"`
	Options   []string `json:"options" form:"options"`
	MaxAnswer int      `json:"max_answer" form:"max_answer"`

	// EndTime agenda o encerramento automático da enquete em um dia e horário
	// específicos. Aceita dois formatos no mesmo campo:
	//
	//   - número: Unix milissegundos desde 1970 (o mesmo valor que Date.now()
	//     devolve no JavaScript) — formato original, que continua intacto;
	//   - string: horário do dia "HH:MM:SS" ou "HH:MM", que precisa vir
	//     acompanhado de end_date; a API combina os dois no fuso de Brasília
	//     e faz o cálculo de milissegundos sozinha.
	//
	// Opcional. Quando não informado, a enquete fica aberta indefinidamente,
	// exatamente como antes — nenhum campo existente muda de comportamento.
	EndTime *PollEndTime `json:"end_time,omitempty" form:"end_time"`

	// EndDate informa o dia do encerramento ("AAAA-MM-DD" ou "DD/MM/AAAA").
	// Só tem efeito junto com end_time em formato de horário e não pode ser
	// combinado com end_time numérico.
	EndDate *string `json:"end_date,omitempty" form:"end_date"`
}

// PollEndTime permite que end_time aceite, no mesmo campo do JSON, um número
// (Unix milissegundos) ou uma string de horário ("HH:MM[:SS]"). Exatamente um
// dos dois fica preenchido, conforme o tipo recebido.
type PollEndTime struct {
	// Millis é preenchido quando end_time veio como número (Unix ms).
	Millis *int64
	// Clock é preenchido quando end_time veio como string "HH:MM[:SS]".
	Clock *string
}

// UnmarshalJSON aceita número (Unix milissegundos) ou string (horário do
// dia); qualquer outro tipo é erro. A validação do conteúdo fica nas
// validations — aqui só se decide o formato recebido.
func (p *PollEndTime) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "null" {
		return nil
	}

	if strings.HasPrefix(trimmed, `"`) {
		var clock string
		if err := json.Unmarshal(data, &clock); err != nil {
			return fmt.Errorf("end_time: %w", err)
		}
		p.Millis, p.Clock = nil, &clock
		return nil
	}

	var millis int64
	if err := json.Unmarshal(data, &millis); err != nil {
		return fmt.Errorf(`end_time must be a number (Unix milliseconds) or a "HH:MM:SS" string`)
	}
	p.Millis, p.Clock = &millis, nil
	return nil
}

// MarshalJSON devolve o campo na mesma forma recebida (número ou string),
// preservando o round-trip do contrato JSON.
func (p PollEndTime) MarshalJSON() ([]byte, error) {
	if p.Clock != nil {
		return json.Marshal(*p.Clock)
	}
	if p.Millis != nil {
		return json.Marshal(*p.Millis)
	}
	return []byte("null"), nil
}

// UnmarshalText dá suporte a form-data: como ali tudo é string, um valor
// puramente numérico é tratado como milissegundos e qualquer outro como
// horário do dia (a validação de conteúdo é a mesma do JSON).
func (p *PollEndTime) UnmarshalText(text []byte) error {
	value := strings.TrimSpace(string(text))
	if value == "" {
		return nil
	}
	if millis, err := strconv.ParseInt(value, 10, 64); err == nil {
		p.Millis, p.Clock = &millis, nil
		return nil
	}
	p.Millis, p.Clock = nil, &value
	return nil
}
