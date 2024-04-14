package twitter

import "encoding/json"

type JsonTreeParser interface {
	ParseMap(map[string]interface{})
	ParseArray([]interface{})
}

func parse(parser JsonTreeParser, a interface{}) {
	switch concreteVal := a.(type) {
	case map[string]interface{}:
		parser.ParseMap(concreteVal)
	case []interface{}:
		parser.ParseArray(concreteVal)
	default:
	}
}

func parseArray(parser JsonTreeParser, anArray []interface{}) {
	for _, val := range anArray {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			parser.ParseMap(concreteVal)
		case []interface{}:
			parser.ParseArray(concreteVal)
		default:
		}
	}
}

func parseMap(parser JsonTreeParser, aMap map[string]interface{}) {
	for _, val := range aMap {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			parser.ParseMap(concreteVal)
		case []interface{}:
			parser.ParseArray(concreteVal)
		default:
		}
	}
}

func hasKey(aMap map[string]interface{}, key string) bool {
	_, ok := aMap[key]
	return ok
}

func tryGetKeyString(aMap map[string]interface{}, key string) (string, bool) {
	val, ok := aMap[key]
	if ok {
		str, ok := val.(string)
		if ok {
			return str, true
		}
	}
	return "", false
}

func tryGetKeyInt(aMap map[string]interface{}, key string) (int, bool) {
	val, ok := aMap[key]
	if ok {
		num, ok := val.(json.Number)
		n, err := num.Int64()
		if err != nil {
			return 0, false
		}
		if ok {
			return int(n), true
		}
	}
	return 0, false

}
