package events

// import (
// 	"encoding/json"
// 	"fmt"
// 	"reflect"
//
// 	"github.com/pkg/errors"
// )
// //
// // func UnpackJSON(b []byte) (Event, error) {
// 	var e struct {
// 		T string `json:"type"`
// 		// P []byte `json:"ev"`
// 	}
// 	err := json.Unmarshal(b, &e)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "can't unmarshal json into event envelope")
// 	}
//
// 	tepy, found := evRegistry[e.T]
// 	if !found {
// 		return nil, errors.Errorf("Unable to find event type %s in event registry")
// 	}
// 	v := reflect.New(tepy).Elem()
// 	// err = json.Unmarshal( &v)
// 	// if err != nil {
// 	// 	return nil, errors.Wrap(err, "can't unmarshal json payload into event type")
// 	// }
// 	fmt.Printf("%#v", v)
// 	return nil, nil
// }
