package repository

var defaultGrammarMD = `# Grammar Rules

## Basic Sentence Structure
- Every Korean sentence **must** end in either a verb, an adjective, or the copula \` + "`이다`" + ` (to be).
- The language relies on two fundamental word orders:
  - **Subject - Object - Verb (SOV)** (e.g., I hamburger eat)
  - **Subject - Adjective (SA)** (e.g., I beautiful)
- *Note for generation:* For this specific lesson, actual verbs and adjectives are not yet conjugated; sentences are built strictly using nouns paired with the copula \` + "`이다`" + `.

## Copula (To Be) — 이다
- **Form:** Attached directly to the preceding noun without a space: \` + "`Noun이다`" + `.
- **Function:** Links a subject to a predicate noun (\` + "`A is B`" + `). It acts similarly to an adjective and **cannot** take an object.
- **Sentence Structure:** \` + "`[Subject + 은/는] [Noun + 이다]" + `
- **Examples:**
  - 나는 남자이다 (I am a man)
  - 저는 선생님이다 (I am a teacher)

## Particles
- **은/는 (Topic Marker):** Appended to a noun to indicate it is the main subject or topic of the sentence.
  - Use \` + "`는`" + ` if the preceding noun ends in a vowel (e.g., 나 becomes 나는, 저 becomes 저는).
  - Use \` + "`은`" + ` if the preceding noun ends in a consonant (e.g., 사람 becomes 사람은, 학생 becomes 학생은).
- **을/를 (Object Marker):** Appended to a noun to indicate it is the object of an action.
  - Use \` + "`를`" + ` after a vowel; use \` + "`을`" + ` after a consonant.
  - *Constraint:* Never use object markers with \` + "`이다`" + `, as it cannot act on an object.

## Determiners & Pronouns (This / That)
- **Determiners (Modifiers):** Words placed directly before nouns to specify location or context. They do not change form:
  - \` + "`이`" + `: This (the object is within reaching distance)
  - \` + "`그`" + `: That (the object is understood from context or a previous sentence)
  - \` + "`저`" + `: That (the object is far away and out of reach)
- **Pronouns:** Formed by combining a determiner directly with the noun \` + "`것`" + ` (thing) with no spacing:
  - \` + "`이것`" + `: This thing (shortened to \` + "`이거`" + ` in casual speech)
  - \` + "`그것`" + `: That thing (shortened to \` + "`그거`" + ` in casual speech)
  - \` + "`저것`" + `: That thing over there (shortened to \` + "`저거`" + ` in casual speech)
- **Examples:**
  - 이것은 책이다 (This is a book)
  - 그 사람은 학생이다 (That person is a student)`

var defaultVocab = []VocabEntry{
	{Korean: "한국", English: "Korea"},
	{Korean: "도시", English: "city"},
	{Korean: "이름", English: "name"},
	{Korean: "저", English: "I (formal)"},
	{Korean: "나", English: "I (informal)"},
	{Korean: "남자", English: "man"},
	{Korean: "여자", English: "woman"},
	{Korean: "이", English: "this"},
	{Korean: "그", English: "that"},
	{Korean: "저", English: "that (far away)"},
	{Korean: "것", English: "thing"},
	{Korean: "이것", English: "this thing"},
	{Korean: "그것", English: "that thing"},
	{Korean: "저것", English: "that thing (far away)"},
	{Korean: "의자", English: "chair"},
	{Korean: "탁자", English: "table"},
	{Korean: "선생님", English: "teacher"},
	{Korean: "침대", English: "bed"},
	{Korean: "집", English: "house"},
	{Korean: "차", English: "car"},
	{Korean: "사람", English: "person"},
	{Korean: "책", English: "book"},
	{Korean: "컴퓨터", English: "computer"},
	{Korean: "나무", English: "tree"},
	{Korean: "소파", English: "sofa"},
	{Korean: "지갑", English: "wallet"},
	{Korean: "방", English: "room"},
	{Korean: "문", English: "door"},
	{Korean: "의사", English: "doctor"},
	{Korean: "학생", English: "student"},
}

func Seed(r Repository) error {
	if err := r.CreatePhase(Phase{Number: 1, Name: "Phase 1"}); err != nil {
		return err
	}
	if err := r.CreateLevel(Level{
		Number:      1,
		PhaseNumber: 1,
		GrammarMD:   defaultGrammarMD,
		ExceptionsMD: "",
	}); err != nil {
		return err
	}
	if err := r.AddVocabulary(1, defaultVocab); err != nil {
		return err
	}
	return nil
}
