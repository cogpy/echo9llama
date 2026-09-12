package consciousness

import "testing"

func TestAnalyzeEmergenceHandlesTenToNineteenMessages(t *testing.T) {
	hub := NewLayerCommunicationHub()
	for index := range 10 {
		messageType := MessageReflection
		if index >= 6 {
			messageType = MessageQuestion
		}
		hub.messageHistory = append(hub.messageHistory, CreateMessage(LayerBasic, LayerReflective, messageType, "evidence", 0.5))
	}

	hub.analyzeEmergence()
	metrics := hub.GetMetrics()
	if got := metrics["emergence_detected"].(uint64); got != 1 {
		t.Fatalf("emergence_detected = %d, want 1", got)
	}
}

func TestGetRecentMessagesRejectsNegativeCount(t *testing.T) {
	hub := NewLayerCommunicationHub()
	hub.messageHistory = append(hub.messageHistory, CreateMessage(LayerBasic, LayerReflective, MessagePattern, "pattern", 0.5))
	if messages := hub.GetRecentMessages(-1); len(messages) != 0 {
		t.Fatalf("negative recent count returned %d messages", len(messages))
	}
}

func TestSendMessageRejectsNilAndKeepsImmutableHistory(t *testing.T) {
	hub := NewLayerCommunicationHub()
	if err := hub.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer hub.Stop()

	if err := hub.SendMessage(nil); err == nil {
		t.Fatal("SendMessage(nil) succeeded")
	}
	message := CreateMessage(LayerBasic, LayerReflective, MessagePattern, "original", 0.5)
	message.Context["topic"] = "wisdom"
	nested := map[string]interface{}{"values": []interface{}{"stable"}}
	message.Context["nested"] = nested
	if err := hub.SendMessage(message); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	message.Content = "mutated"
	message.Context["topic"] = "changed"
	nested["values"].([]interface{})[0] = "mutated"

	history := hub.GetRecentMessages(1)
	if len(history) != 1 || history[0].Content != "original" || history[0].Context["topic"] != "wisdom" {
		t.Fatalf("history did not preserve admitted snapshot: %+v", history)
	}
	if got := history[0].Context["nested"].(map[string]interface{})["values"].([]interface{})[0]; got != "stable" {
		t.Fatalf("nested context did not preserve admitted snapshot: %v", got)
	}
	history[0].Content = "reader mutation"
	history[0].Context["nested"].(map[string]interface{})["values"].([]interface{})[0] = "reader mutation"
	if reread := hub.GetRecentMessages(1); reread[0].Content != "original" {
		t.Fatalf("reader mutated internal history: %+v", reread[0])
	}
	if got := hub.GetRecentMessages(1)[0].Context["nested"].(map[string]interface{})["values"].([]interface{})[0]; got != "stable" {
		t.Fatalf("reader mutated nested internal history: %v", got)
	}

	unsupported := CreateMessage(LayerBasic, LayerReflective, MessagePattern, "unsafe context", 0.5)
	unsupported.Context["channel"] = make(chan struct{})
	if err := hub.SendMessage(unsupported); err == nil {
		t.Fatal("non-serializable context should fail closed")
	}
}
