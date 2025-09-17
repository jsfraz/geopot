package models

import "encoding/json"

type WSMessage struct {
	Type WSMessageType `json:"type"`
	Data Connection    `json:"data"`
}

func NewWSMessage(messageType WSMessageType, connection Connection) *WSMessage {
	return &WSMessage{
		Type: messageType,
		Data: connection,
	}
}

func (w *WSMessage) MarshalBinary() ([]byte, error) {
	return json.Marshal(w)
}
