// +build ignore
package events

// type envelope struct {
// 	ev Event
// 	t  string
// }
//
// func (e envelope) MarshalJSON() ([]byte, error) {
//
// 	var payload []byte
// 	payload, err := json.Marshal(e.ev)
// 	if err != nil {
// 		return payload, errors.Wrap(err, "can't marshal ev.Event (%#v) as json to envelop")
// 	}
//
// 	ev := struct {
// 		Type    string `json:"type"`
// 		Payload string `json:"payload"`
// 	}{
// 		reflect.TypeOf(e.ev).Elem().Name(),
// 		string(payload),
// 	}
//
// 	return json.Marshal(ev)
// }
