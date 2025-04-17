package hw03frequencyanalysis

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Change to true if needed.
var taskWithAsteriskIsCompleted = true

var text = `Как видите, он  спускается  по  лестнице  вслед  за  своим
	другом   Кристофером   Робином,   головой   вниз,  пересчитывая
	ступеньки собственным затылком:  бум-бум-бум.  Другого  способа
	сходить  с  лестницы  он  пока  не  знает.  Иногда ему, правда,
		кажется, что можно бы найти какой-то другой способ, если бы  он
	только   мог   на  минутку  перестать  бумкать  и  как  следует
	сосредоточиться. Но увы - сосредоточиться-то ему и некогда.
		Как бы то ни было, вот он уже спустился  и  готов  с  вами
	познакомиться.
	- Винни-Пух. Очень приятно!
		Вас,  вероятно,  удивляет, почему его так странно зовут, а
	если вы знаете английский, то вы удивитесь еще больше.
		Это необыкновенное имя подарил ему Кристофер  Робин.  Надо
	вам  сказать,  что  когда-то Кристофер Робин был знаком с одним
	лебедем на пруду, которого он звал Пухом. Для лебедя  это  было
	очень   подходящее  имя,  потому  что  если  ты  зовешь  лебедя
	громко: "Пу-ух! Пу-ух!"- а он  не  откликается,  то  ты  всегда
	можешь  сделать вид, что ты просто понарошку стрелял; а если ты
	звал его тихо, то все подумают, что ты  просто  подул  себе  на
	нос.  Лебедь  потом  куда-то делся, а имя осталось, и Кристофер
	Робин решил отдать его своему медвежонку, чтобы оно не  пропало
	зря.
		А  Винни - так звали самую лучшую, самую добрую медведицу
	в  зоологическом  саду,  которую  очень-очень  любил  Кристофер
	Робин.  А  она  очень-очень  любила  его. Ее ли назвали Винни в
	честь Пуха, или Пуха назвали в ее честь - теперь уже никто  не
	знает,  даже папа Кристофера Робина. Когда-то он знал, а теперь
	забыл.
		Словом, теперь мишку зовут Винни-Пух, и вы знаете почему.
		Иногда Винни-Пух любит вечерком во что-нибудь поиграть,  а
	иногда,  особенно  когда  папа  дома,  он больше любит тихонько
	посидеть у огня и послушать какую-нибудь интересную сказку.
		В этот вечер...`

func noWordsTests(t *testing.T) {
	t.Helper()
	t.Run("String with only spaces", func(t *testing.T) { require.Len(t, Top10("             "), 0) })
	t.Run("String with only spaces, tabs and newlines", func(t *testing.T) {
		require.Len(t, Top10("    \t      \n    "), 0)
	})
	t.Run("String with only spaces and punctuation", func(t *testing.T) {
		source := "      ,        .         ⸻      "
		require.Len(t, Top10(source), 0)
	})
}

func singleWordTests(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name     string
		source   string
		expected []string
	}{
		{"Common word", "Word", []string{"word"}},
		{"Repeated word, different cases", "Word word", []string{"word"}},
		{"A word consisting of punctuation marks", ".!-?¡", []string{".!-?¡"}},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			got := Top10(tC.source)
			require.Equal(t, tC.expected, got)
		})
	}
}

func orderTests(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name     string
		source   string
		expected []string
	}{
		{"Several words, same frequency", "f e d c a b", []string{"a", "b", "c", "d", "e", "f"}},
		{
			"Mixed frequencies, lexicographical order",
			"never gonna give you up never gonna let you down",
			[]string{"gonna", "never", "you", "down", "give", "let", "up"},
		},
		{"Several repeating words, only few unique", "a b a b a b a b a b c d e c", []string{"a", "b", "c", "d", "e"}},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			got := Top10(tC.source)
			require.Equal(t, tC.expected, got)
		})
	}
}

func limitTests(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name     string
		source   string
		expected []string
	}{
		{
			"15 unique words, same frequency",
			"a b c d e f g h i j k l m n o",
			[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		},
		{
			"10 words with the same frequency, a few more with lower frequency",
			"a a b b c c d d e e f f g g h h i i j j k l m n o",
			[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		},
		{"Stress test", strings.Repeat("word ", 1000), []string{"word"}},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			got := Top10(tC.source)
			require.Equal(t, tC.expected, got)
		})
	}
}

func punctuationTests(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name     string
		source   string
		expected []string
	}{
		{"Whole word consists of punctuation", "      ,,        .    ---       ⸻      ", []string{",,", "---"}},
		{"Mixed short words", ",a a, ,a, ,,a a,, ,,a,, .b⸻c.", []string{"a", ",,a", ",,a,,", "a,,", "b⸻c"}},
		{
			"Mixed long words",
			",aaa aaa, ,aaa, ,,aaa aaa,, ,,aaa,, .bbb⸻ccc.",
			[]string{"aaa", ",,aaa", ",,aaa,,", "aaa,,", "bbb⸻ccc"},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			got := Top10(tC.source)
			require.Equal(t, tC.expected, got)
		})
	}
}

func additionalTests(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name     string
		source   string
		expected []string
	}{
		{"Digits and special characters", "1 1 2 @ @ @ 3", []string{"1", "2", "3"}},
		{"Unicode chars", "世界 オラ オラ オラ オラ オラ オラ オラ オラ オラ オラ ³ ७ ७ Ⅸ", []string{"オラ", "७", "³", "ⅸ", "世界"}},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			got := Top10(tC.source)
			require.Equal(t, tC.expected, got)
		})
	}
}

func TestTop10(t *testing.T) {
	t.Run("no words in empty string", func(t *testing.T) {
		require.Len(t, Top10(""), 0)
	})

	t.Run("positive test", func(t *testing.T) {
		if taskWithAsteriskIsCompleted {
			expected := []string{
				"а",         // 8
				"он",        // 8
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"в",         // 4
				"его",       // 4
				"если",      // 4
				"кристофер", // 4
				"не",        // 4
			}
			require.Equal(t, expected, Top10(text))
		} else {
			expected := []string{
				"он",        // 8
				"а",         // 6
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"-",         // 4
				"Кристофер", // 4
				"если",      // 4
				"не",        // 4
				"то",        // 4
			}
			require.Equal(t, expected, Top10(text))
		}
	})

	t.Run("No words", func(t *testing.T) { noWordsTests(t) })
	t.Run("Single word", func(t *testing.T) { singleWordTests(t) })
	t.Run("Output order", func(t *testing.T) { orderTests(t) })
	t.Run("Top-10 and overall limit", func(t *testing.T) { limitTests(t) })
	t.Run("Punctuation cases", func(t *testing.T) { punctuationTests(t) })
	t.Run("Additional tests", func(t *testing.T) { additionalTests(t) })
}
