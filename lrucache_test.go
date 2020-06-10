/*
Copyright 2020 Daniël Franke

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lrucache_test

import (
	"fmt"
	"testing"
	"time"

	lrucache "github.com/ainmosni/golang-lrucache"
)

// factorial is a simple heavy function that we use to test the cache with.
func factorial(n uint64) uint64 {
	// We can't get a factorial for a negative number.
	if n > 0 {
		var result uint64 = 1
		for i := uint64(1); i <= n; i++ {
			result *= i
		}
		return result
	}
	return 0
}

func TestCacheTime(t *testing.T) {
	testcases := []struct {
		Input  uint64
		Output uint64
	}{
		{
			Input:  2,
			Output: 2,
		},
		{
			Input:  3,
			Output: 6,
		},
		{
			Input:  4,
			Output: 24,
		},
		{
			Input:  50,
			Output: 15188249005818642432,
		},
		{
			Input:  53,
			Output: 13175843659825807360,
		},
		{
			Input:  60,
			Output: 9727775195120271360,
		},
		{
			Input:  65,
			Output: 9223372036854775808,
		},
	}

	fc := lrucache.New(5, factorial)
	for _, tc := range testcases {
		t.Run(fmt.Sprintf("testing cache %d", tc.Input), func(t *testing.T) {
			coldStart := time.Now()
			res := fc.Call(tc.Input) //nolint
			if res != tc.Output {    //nolint
				t.Fatalf("Expected %d when calculating factorial %d, got %d", tc.Output, tc.Input, res) //nolint
			}
			coldTime := time.Since(coldStart)
			warmStart := time.Now()

			res = fc.Call(tc.Input) //nolint
			if res != tc.Output {   //nolint
				t.Fatalf("Expected %d when calculating factorial %d, got %d", tc.Output, tc.Input, res) //nolint
			}
			warmTime := time.Since(warmStart)
			if warmTime >= coldTime {
				t.Fatalf("Expected cached response to be faster, cached time: %d, cold time: %d", warmTime, coldTime)
			}
		})
	}
}

func TestExpiry(t *testing.T) {
	t.Run("test cache expiry.", func(t *testing.T) {
		fc := lrucache.New(2, factorial)
		_ = fc.Call(1)
		if _, ok := fc.ResponsesLookup[1]; !ok {
			t.Fatal("Couldn't find `1` in cache")
		}
		_ = fc.Call(2)
		_ = fc.Call(3)
		if _, ok := fc.ResponsesLookup[1]; ok {
			t.Fatal("Not evicted: `1`")
		}
	})
}

func TestRecency(t *testing.T) {
	t.Run("test cache expiry.", func(t *testing.T) {
		fc := lrucache.New(3, factorial)
		_ = fc.Call(1)
		if fc.ResponsesList.Front().Value.(*lrucache.Entry).Key != 1 {
			t.Fatal("`1` isn't in front of the cache.")
		}
		_ = fc.Call(2)
		if fc.ResponsesList.Front().Value.(*lrucache.Entry).Key != 2 {
			t.Fatal("`2` isn't in front of the cache.")
		}
		_ = fc.Call(1)
		if fc.ResponsesList.Front().Value.(*lrucache.Entry).Key != 1 {
			t.Fatal("`1` isn't in front of the cache.")
		}
	})
}
