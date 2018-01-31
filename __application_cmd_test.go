// +build ignore
package main

//
// YAML PARSER IS BROKEN
//
// IT IS NOT POSSIBLE TO UNMARSHAL UNKNOWN YAML BECAUSE IT IS ALL OR
// NOTHING, THERE IS NO WAY TO PEEK INTO LEVEL ONE AND LOOK AT THE VALUES
// AND SWITCH TO DO SOMETHING DIFFERENT IN CASE OF APPLY vs. START SESSION

// import (
// 	"fmt"
// 	"testing"
//
// 	yaml "gopkg.in/yaml.v2"
// )
//
// type ApplicationCmd interface {
// 	Meth() string
// 	Ctxt() string
// 	Session() string
// }
//
// type SerializedCmd struct {
// 	meth string
// 	sesh string
// 	ctxt string
// }
//
// func (sc *SerializedCmd) Meth() string { return sc.meth }
// func (sc *SerializedCmd) Ctxt() string { return sc.sesh }
// func (sc *SerializedCmd) Sesh() string { return sc.ctxt }
//
// func (sc *SerializedCmd) UnmarshalYAML(unmarshal func(interface{}) error) error {
// 	s := make(map[string]interface{})
// 	err := unmarshal(s)
// 	fmt.Println(s)
// 	fmt.Println(err)
// 	if _, ok := s["Apply"]; ok {
// 		fmt.Println("Found a command to Apply")
// 	}
// 	if _, ok := s["StartSession"]; ok {
// 		fmt.Println("Found a command to StartSession")
// 	}
// 	return nil
// }
//
// func Test_SerializedCmd_UnmarshalYAML(t *testing.T) {
// 	testCases := []struct {
// 		yaml string
// 		err  error
// 		cmd  SerializedCmd
// 	}{
// 		{`---
//       - Apply:
//         ctxt: null
//         sesh: null
//         meth: AllowCreationOfNewIdentities`,
// 			nil,
// 			SerializedCmd{meth: "AllowCreationOfNewIdentities"},
// 		},
// 	}
// 	for i, tc := range testCases {
// 		var (
// 			err error
// 			sc  SerializedCmd = SerializedCmd{}
// 		)
// 		err = yaml.Unmarshal([]byte(tc.yaml), &sc)
// 		if err != tc.err {
// 			t.Fatalf("[%d] mismatched err value got=%s wanted=%s for input %s\n", i, err, tc.err, tc.yaml)
// 		}
// 	}
// }
