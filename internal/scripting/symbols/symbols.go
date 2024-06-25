package symbols

import "reflect"

var Symbols = make(map[string]map[string]reflect.Value)

func Clone() map[string]map[string]reflect.Value {
	clone := make(map[string]map[string]reflect.Value)
	for k1, v1 := range Symbols {
		nv1 := make(map[string]reflect.Value)
		for k2, v2 := range v1 {
			nv1[k2] = v2
		}
		clone[k1] = nv1
	}
	return clone
}
