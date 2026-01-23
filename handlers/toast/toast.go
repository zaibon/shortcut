package toast

import (
	"encoding/json"
	"net/http"
)

const (
	TypeInfo    = "info"
	TypeSuccess = "success"
	TypeWarning = "warning"
	TypeError   = "error"

	LevelFlash = "flash"
)

type Toast struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

func New(level, title, message string, customWrapper ...string) Toast {
	return Toast{
		Level:   LevelFlash,
		Message: message,
		Type:    level,
	}
}

func Info(title, message string, customWrapper ...string) Toast {
	return New(TypeInfo, title, message)
}

func Success(w http.ResponseWriter, title, message string, customWrapper ...string) {
	New(TypeSuccess, title, message).SetHXTriggerHeader(w)
}

func Warning(w http.ResponseWriter, title, message string, customWrapper ...string) {
	New(TypeWarning, title, message).SetHXTriggerHeader(w)
}

func Danger(w http.ResponseWriter, title, message string, customWrapper ...string) {
	New(TypeError, title, message).SetHXTriggerHeader(w)
}

func (t Toast) jsonify() (string, error) {
	eventMap := map[string]Toast{}
	eventMap["showMessage"] = t
	jsonData, err := json.Marshal(eventMap)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (t Toast) SetHXTriggerHeader(w http.ResponseWriter) {
	jsonData, _ := t.jsonify()
	w.Header().Set("HX-Trigger", jsonData)
}
