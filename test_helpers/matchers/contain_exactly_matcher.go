package matchers

import (
	"fmt"
	"reflect"

	//	"errors"
	"errors"
	"github.com/onsi/gomega/types"
)

type containExactlyMatcher struct {
	expected interface{}
}

func ContainExactly(expected interface{}) types.GomegaMatcher {
	return &containExactlyMatcher{expected: expected}
}

func (matcher *containExactlyMatcher) Match(actual interface{}) (success bool, err error) {

	if !isArrayOrSlice(matcher.expected) || !isArrayOrSlice(actual) {
		return false, errors.New("Matcher can only take an array or slice")
	}

	actualValue := reflect.ValueOf(actual)
	expectedValue := reflect.ValueOf(matcher.expected)

	if actualValue.Len() != expectedValue.Len() {
		return false, nil
	}

	actualValueOccurrences := make(map[interface{}]int)
	for i := 0; i < actualValue.Len(); i++ {
		actualValueOccurrences[fmt.Sprintf("%#v", actualValue.Index(i).Interface())]++
	}

	expectedValueOccurrences := make(map[interface{}]int)
	for i := 0; i < expectedValue.Len(); i++ {
		expectedValueOccurrences[fmt.Sprintf("%#v", expectedValue.Index(i).Interface())]++
	}

	return reflect.DeepEqual(expectedValueOccurrences, actualValueOccurrences), nil
}

func (matcher *containExactlyMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %#v\n to contain exactly: %#v\n but it did not.", actual, matcher.expected)
}

func (matcher *containExactlyMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %#v\n not to contain exactly: %#v\n but it did!", actual, matcher.expected)
}

func isArrayOrSlice(a interface{}) bool {
	if a == nil {
		return false
	}
	switch reflect.TypeOf(a).Kind() {
	case reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}
