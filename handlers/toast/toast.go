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
	Level         string `json:"level"`
	Title         string `json:"title"`
	Message       string `json:"message"`
	CustomWrapper string `json:"customWrapper"`
}

func New(level, title, message string, customWrapper ...string) Toast {
	t := Toast{
		Level:   level,
		Title:   title,
		Message: message,
	}
	if len(customWrapper) > 0 {
		t.CustomWrapper = customWrapper[0]
	}
	return t
}

func Info(title, message string, customWrapper ...string) Toast {
	return New(INFO, title, message)
}

func Success(w http.ResponseWriter, title, message string, customWrapper ...string) {
	New(SUCCESS, title, message).SetHXTriggerHeader(w)
}

func Warning(w http.ResponseWriter, title, message string, customWrapper ...string)  {
	 New(WARNING, title, message).SetHXTriggerHeader(w)
}

func Danger(w http.ResponseWriter, title, message string, customWrapper ...string)  {
	 New(DANGER, title, message).SetHXTriggerHeader(w)
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
