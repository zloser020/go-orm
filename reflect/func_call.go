package reflect

import "reflect"

func IterateFunc(entity any) (map[string]FuncInfo, error) {
	typ := reflect.TypeOf(entity)
	numMethod := typ.NumMethod()
	res := make(map[string]FuncInfo, numMethod)
	for i := 0; i < numMethod; i++ {
		method := typ.Method(i)
		fn := method.Func

		numIn := method.Type.NumIn()
		input := make([]reflect.Type, 0, numIn)
		input = append(input, reflect.TypeOf(entity))

		inputValues := make([]reflect.Value, 0, numIn)
		inputValues = append(inputValues, reflect.ValueOf(entity))

		for j := 1; j < numIn; j++ {
			fnInputType := method.Type.In(j)
			input = append(input, fnInputType)
			inputValues = append(inputValues, reflect.Zero(fnInputType))
		}

		numOut := method.Type.NumOut()
		output := make([]reflect.Type, 0, numOut)
		for j := 0; j < numOut; j++ {
			output = append(output, fn.Type().Out(j))
		}
		resValues := fn.Call(inputValues)
		result := make([]any, 0, len(resValues))
		for _, res := range resValues {
			result = append(result, res.Interface())
		}
		res[method.Name] = FuncInfo{
			Name:        method.Name,
			InputTypes:  input,
			OutputTypes: output,
			Result:      result,
		}
	}
	return res, nil
}

type FuncInfo struct {
	Name        string
	InputTypes  []reflect.Type
	OutputTypes []reflect.Type
	Result      []any
}
