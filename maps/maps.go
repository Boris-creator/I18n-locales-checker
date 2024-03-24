package maps

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type MapDiff[K, V comparable] struct {
	AddedFields    map[K]V
	ModifiedFields map[K]V
	DroppedFields  map[K]V
}

func (diff MapDiff[K, V]) IsEmpty() bool {
	return len(diff.AddedFields) == 0 && len(diff.ModifiedFields) == 0 && len(diff.DroppedFields) == 0
}
func (diff MapDiff[K, V]) String() string {
	if diff.IsEmpty() {
		return ""
	}
	m := make(map[string]map[K]V, 3)
	if len(diff.AddedFields) != 0 {
		m["added"] = diff.AddedFields
	}
	encoded, _ := json.Marshal(m)
	return string(encoded)
}

func newMapDiff[K, V comparable]() MapDiff[K, V] {
	return MapDiff[K, V]{
		map[K]V{},
		map[K]V{},
		map[K]V{},
	}
}

func Accessible(unknown any) bool {
	return reflect.TypeOf(unknown).Kind() == reflect.Map
}

func ToMap[K comparable, V any](unknown any) (map[K]V, error) {
	if !Accessible(unknown) {
		return nil, fmt.Errorf("not a map")
	}
	mapType := reflect.ValueOf(unknown)
	p := make(map[K]V, mapType.Len())
	for _, key := range mapType.MapKeys() {
		typedKey, ok := key.Interface().(K)
		if !ok {
			return nil, fmt.Errorf("cannot convert key")
		}
		typedValue, ok := mapType.MapIndex(key).Interface().(V)
		if !ok {
			return nil, fmt.Errorf("cannot convert value")
		}
		p[typedKey] = typedValue
	}
	return p, nil
}

func Merge[K comparable, V any](target, source map[K]V) map[K]V {
	for k, v := range source {
		target[k] = v
	}
	return target
}

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func Divide[K comparable, V any](m map[K]V) ([]K, []V) {
	keys := make([]K, 0, len(m))
	values := make([]V, 0, len(m))
	for k, v := range m {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
}

func Map[K, K1 comparable, V, V1 any](m map[K]V, f func(K, V) (K1, V1)) map[K1]V1 {
	newMap := make(map[K1]V1, len(m))
	for key, value := range m {
		newKey, newValue := f(key, value)
		newMap[newKey] = newValue
	}
	return newMap
}

func Dot[K comparable, V any](m map[K]any) map[string]V {
	result := make(map[string]V, len(m))
	for k, v := range m {
		if Accessible(v) {
			p, _ := ToMap[K, any](v)
			nested := Dot[K, V](map[K]any(p))
			Merge(result, PrependKeysWith(nested, fmt.Sprintf("%v.", k)))
			continue
		}
		result[fmt.Sprint(k)] = v.(V)
	}
	return result
}

func Set[V any](m map[string]any, key string, value V) map[string]any {
	keys := strings.Split(key, ".")
	if len(keys) == 1 {
		m[keys[0]] = value
		return m
	}
	if value, ok := m[keys[0]]; !ok || !Accessible(value) {
		m[keys[0]] = make(map[string]any)
	}
	Set(m[keys[0]].(map[string]any), strings.Join(keys[1:], "."), value)
	return m
}

func Get[V any](m map[string]any, key string, defaultValue V) V {
	keys := strings.Split(key, ".")
	value, ok := m[keys[0]]
	if !ok {
		return defaultValue
	}
	if len(keys) == 1 {
		return value.(V)
	}
	m, _ = ToMap[string, any](value)
	return Get(m, strings.Join(keys[1:], "."), defaultValue)
}

func Undot[V any](m map[string]V) map[string]any {
	newMap := make(map[string]any, len(m))
	sortedKeys := Keys(m)
	sort.SliceStable(sortedKeys, func(i, j int) bool {
		return strings.HasPrefix(sortedKeys[j], sortedKeys[i])
	})
	for _, k := range sortedKeys {
		Set(newMap, k, m[k])
	}
	return newMap
}

func PrependKeysWith[K comparable, V any](m map[K]V, prefix string) map[string]V {
	prefixed := make(map[string]V, len(m))
	for k, v := range m {
		prefixed[prefix+fmt.Sprint(k)] = v
	}
	return prefixed
}

func Diff[V, K comparable](m1, m2 map[K]V) MapDiff[K, V] {
	diff := newMapDiff[K, V]()
	for k, v := range m1 {
		v1, exists := m2[k]
		if !exists {
			diff.AddedFields[k] = v
			continue
		}
		if v1 != v {
			diff.ModifiedFields[k] = v
		}
	}
	for k, v := range m2 {
		_, exists := m1[k]
		if !exists {
			diff.AddedFields[k] = v
		}
	}
	return diff
}
