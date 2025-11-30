package adaptive

import "strings"

// Mongo decide dinamicamente o backend com base em tags, nome de tipo
// ou metadado GetType() dos models (já implementado em todos).
func Mongo(v any) bool {
	if v == nil {
		return false
	}
	if t, ok := v.(interface{ GetType() string }); ok {
		tp := strings.ToLower(t.GetType())
		return strings.HasPrefix(tp, "mongo") || strings.Contains(tp, "nosql")
	}
	return false
}
