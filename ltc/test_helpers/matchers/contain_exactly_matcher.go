package matchers

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/onsi/gomega/types"
)

type containExactlyMatcher struct {
	expected interface{}
}

func ContainExactly(expected interface{}) types.GomegaMatcher {
	return &containExactlyMatcher{expected: expected}
}

func (matcher *containExactlyMatcher) Match(actual interface{}) (success bool, err error) {

	if !isArraySliceMap(matcher.expected) || !isArraySliceMap(actual) {
		return false, errors.New("Matcher can only take an array, slice or map")
	}

	expectedValueOccurrences := calculateOccurrencesMap(matcher.expected)
	actualValueOccurrences := calculateOccurrencesMap(actual)

	return reflect.DeepEqual(expectedValueOccurrences, actualValueOccurrences), nil
}

func calculateOccurrencesMap(actualOrExpected interface{}) map[interface{}]int {
	value := reflect.ValueOf(actualOrExpected)
	occurrences := make(map[interface{}]int)
	var keys []reflect.Value
	if isMap(actualOrExpected) {
		keys = value.MapKeys()
	}

	for i := 0; i < value.Len(); i++ {
		var valueToHash interface{}
		if isMap(actualOrExpected) {
			valueToHash = value.MapIndex(keys[i]).Interface()
		} else {
			valueToHash = value.Index(i).Interface()
		}
		occurrences[fmt.Sprintf("%#v", valueToHash)]++
	}

	return occurrences
}

func (matcher *containExactlyMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %#v\n to contain exactly: %#v\n but it did not.", actual, matcher.expected)
}

func (matcher *containExactlyMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %#v\n not to contain exactly: %#v\n but it did!", actual, matcher.expected)
}

func isArraySliceMap(a interface{}) bool {
	if a == nil {
		return false
	}
	switch reflect.TypeOf(a).Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return true
	default:
		return false
	}
}

func isMap(a interface{}) bool {
	//	if a == nil {
	//		return false
	//	}
	return reflect.TypeOf(a).Kind() == reflect.Map
}
