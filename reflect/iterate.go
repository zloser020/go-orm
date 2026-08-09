package reflect

import "reflect"

func IterateArrayOrSlice(entity any) ([]any, error) {
	val := reflect.ValueOf(entity)
	res := make([]any, 0, val.Len())
	for i := 0; i < val.Len(); i++ {
		element := val.Index(i)
		res = append(res, element.Interface())
	}
	return res, nil
}

func IterateMap(entity any) ([]any, []any, error) {
	val := reflect.ValueOf(entity)
	resKeys := make([]any, 0, val.Len())
	resVals := make([]any, 0, val.Len())

	iter := val.MapRange()
	for iter.Next() {
		resKeys = append(resKeys, iter.Key().Interface())
		resVals = append(resVals, iter.Value().Interface())
	}

	//keys := val.MapKeys()
	//for _, key := range keys {
	//	element := val.MapIndex(key)
	//	resKeys = append(resKeys, key.Interface())
	//	resVals = append(resVals, element.Interface())
	//}
	return resKeys, resVals, nil
}
