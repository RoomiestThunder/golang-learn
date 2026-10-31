---
title: Частота слов
level: medium
---
@@ statement
1. `WordFreq(text string) map[string]int` — сколько раз встречается каждое слово. Слова приводятся к нижнему регистру; знаки препинания (`.,!?;:"'()-—`) отбрасываются с краёв слов; разделитель — пробельные символы (`strings.Fields`). Пустые «слова» не считаются.
2. `TopN(freq map[string]int, n int) []string` — `n` самых частых слов. Порядок: по убыванию частоты, при равенстве — по алфавиту. Если слов меньше `n` — вернуть все.

Подсказки: `strings.ToLower`, `strings.Trim`/`strings.TrimFunc`, `unicode.IsLetter`, `sort.Slice` или `slices.SortFunc`.
@@ starter
package solution

func WordFreq(text string) map[string]int {
	return nil
}

func TopN(freq map[string]int, n int) []string {
	return nil
}
@@ tests
package solution

import (
	"reflect"
	"slices"
	"testing"
)

func TestWordFreq(t *testing.T) {
	got := WordFreq("Go, go! GO... это Go-язык? Да, да.")
	want := map[string]int{"go": 3, "это": 1, "go-язык": 1, "да": 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if got := WordFreq("  ... !!! "); len(got) != 0 {
		t.Errorf("только знаки препинания: %v", got)
	}
}

func TestTopN(t *testing.T) {
	freq := map[string]int{"b": 2, "a": 2, "c": 5, "d": 1}
	if got := TopN(freq, 3); !slices.Equal(got, []string{"c", "a", "b"}) {
		t.Errorf("TopN(3) = %v", got)
	}
	if got := TopN(freq, 10); !slices.Equal(got, []string{"c", "a", "b", "d"}) {
		t.Errorf("TopN(10) = %v", got)
	}
	if got := TopN(freq, 0); len(got) != 0 {
		t.Errorf("TopN(0) = %v", got)
	}
}
@@ solution
package solution

import (
	"sort"
	"strings"
	"unicode"
)

func WordFreq(text string) map[string]int {
	freq := map[string]int{}
	for _, w := range strings.Fields(strings.ToLower(text)) {
		w = strings.TrimFunc(w, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
		if w != "" {
			freq[w]++
		}
	}
	return freq
}

func TopN(freq map[string]int, n int) []string {
	words := make([]string, 0, len(freq))
	for w := range freq {
		words = append(words, w)
	}
	sort.Slice(words, func(i, j int) bool {
		if freq[words[i]] != freq[words[j]] {
			return freq[words[i]] > freq[words[j]]
		}
		return words[i] < words[j]
	})
	if n < len(words) {
		words = words[:max(n, 0)]
	}
	return words
}
@@ hint
Для `TopN`: собери ключи в слайс и отсортируй с компаратором по двум критериям.
