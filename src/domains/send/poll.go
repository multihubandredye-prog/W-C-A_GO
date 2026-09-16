package send

type PollRequest struct {
	BaseRequest
	Question  string   `json:"question" form:"question"`
	Options   []string `json:"options" form:"options"`
	MaxAnswer int      `json:"max_answer" form:"max_answer"`

	// EndTime agenda o encerramento automático da enquete em um dia e horário
	// específicos. Valor em milissegundos desde 1970 (Unix ms — o mesmo valor que
	// Date.now() devolve no JavaScript).
	//
	// Opcional. Quando não informado, a enquete fica aberta indefinidamente,
	// exatamente como antes — nenhum campo existente muda de comportamento.
	EndTime *int64 `json:"end_time,omitempty" form:"end_time"`
}
