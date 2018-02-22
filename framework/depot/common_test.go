package depot

// func Test_Depot(t *testing.T) {
// 	depots := map[string]types.Depot{
// 		"in-memory": memory.NewEmptyDepot(),
// 		"redis":     redis.NewDepot(),
// 	}
// 	for name, depot := range depots {
// 		t.Run(name, func(t *testing.T) {
//
// 			t.Run("validation", func(t *testing.T) {
// 				t.Run("must refuse to store for paths not including an ID part, except _", func(t *testing.T) {
// 					t.Skip("not implemented yet")
// 				})
//
// 				t.Run("must allow events to survive a roundtrip of storage (incl args)", func(t *testing.T) {
// 					t.Skip("not implemented yet")
// 				})
// 			})
//
// 			t.Run("static-queries", func(t *testing.T) {
// 				t.Run("must allow lookup by verbatim path", func(t *testing.T) {
// 					t.Skip("not implemented yet")
// 				})
//
// 				t.Run("must allow lookup by globbing", func(t *testing.T) {
// 					t.Skip("not implemented yet")
// 				})
// 			})
//
// 			t.Run("rehydrate", func(t *testing.T) {
// 				t.Skip("must be able to rehydrate things")
// 			})
//
// 		})
// 	}
// }
