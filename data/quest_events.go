package data

import "log"

// Quest events are stored one node each, keyed by name, with the whole
// definition as a JSON string; objects.QuestEvent owns the format.

func LoadQuestEvents() map[string]string {
	events := make(map[string]string)
	results, err := execRead("MATCH (q:quest_event) RETURN {name: q.name, definition: q.definition}", nil)
	if err != nil {
		log.Println(err)
		return events
	}
	for _, row := range results {
		event := row.Values[0].(map[string]interface{})
		name, _ := event["name"].(string)
		definition, _ := event["definition"].(string)
		if name != "" {
			events[name] = definition
		}
	}
	return events
}

func SaveQuestEvent(name string, definition string) bool {
	_, err := execWrite(
		"MERGE (q:quest_event {name: $name}) SET q.definition = $definition",
		map[string]interface{}{
			"name":       name,
			"definition": definition,
		},
	)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func DeleteQuestEvent(name string) bool {
	_, err := execWrite(
		"MATCH (q:quest_event {name: $name}) DELETE q",
		map[string]interface{}{
			"name": name,
		},
	)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
