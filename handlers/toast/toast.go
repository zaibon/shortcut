package toast

import (
	"encoding/json"
	"net/http"
)

const (
	INFO    = "info"
	SUCCESS = "success"
	WARNING = "warning"
	DANGER  = "danger"
)

type Toast struct {
	Level   string `json:"level"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

func New(level string, title, message string) Toast {
	return Toast{level, title, message}
}

func Info(title, message string) Toast {
	return New(INFO, title, message)
}

func Success(w http.ResponseWriter, title, message string) {
	New(SUCCESS, title, message).SetHXTriggerHeader(w)
}

func Warning(title, message string) Toast {
	return New(WARNING, title, message)
}

func Danger(title, message string) Toast {
	return New(DANGER, title, message)
}

func (t Toast) jsonify() (string, error) {
	eventMap := map[string]Toast{}
	eventMap["makeToast"] = t
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
