INSERT INTO phases (number, name) VALUES (1, 'Phase 1');

INSERT INTO levels (number, phase_number, grammar_md, exceptions_md)
VALUES (
  1,
  1,
'# Grammar Rules

## Basic Sentence Structure
- Every Korean sentence **must** end in either a verb, an adjective, or the copula `이다` (to be).
- The language relies on two fundamental word orders:
  - **Subject - Object - Verb (SOV)** (e.g., I hamburger eat)
  - **Subject - Adjective (SA)** (e.g., I beautiful)
- *Note for generation:* For this specific lesson, actual verbs and adjectives are not yet conjugated; sentences are built strictly using nouns paired with the copula `이다`.

## Copula (To Be) — 이다
- **Form:** Attached directly to the preceding noun without a space: `Noun이다`.
- **Function:** Links a subject to a predicate noun (`A is B`). It acts similarly to an adjective and **cannot** take an object.
- **Sentence Structure:** `[Subject + 은/는] [Noun + 이다]`
- **Examples:**
  - 나는 남자이다 (I am a man)
  - 저는 선생님이다 (I am a teacher)

## Particles
- **은/는 (Topic Marker):** Appended to a noun to indicate it is the main subject or topic of the sentence.
  - Use `는` if the preceding noun ends in a vowel (e.g., 나 becomes 나는, 저 becomes 저는).
  - Use `은` if the preceding noun ends in a consonant (e.g., 사람 becomes 사람은, 학생 becomes 학생은).
- **을/를 (Object Marker):** Appended to a noun to indicate it is the object of an action.
  - Use `를` after a vowel; use `을` after a consonant.
  - *Constraint:* Never use object markers with `이다`, as it cannot act on an object.

## Determiners & Pronouns (This / That)
- **Determiners (Modifiers):** Words placed directly before nouns to specify location or context. They do not change form:
  - `이`: This (the object is within reaching distance)
  - `그`: That (the object is understood from context or a previous sentence)
  - `저`: That (the object is far away and out of reach)
- **Pronouns:** Formed by combining a determiner directly with the noun `것` (thing) with no spacing:
  - `이것`: This thing (shortened to `이거` in casual speech)
  - `그것`: That thing (shortened to `그거` in casual speech)
  - `저것`: That thing over there (shortened to `저거` in casual speech)
- **Examples:**
  - 이것은 책이다 (This is a book)
  - 그 사람은 학생이다 (That person is a student)',
  NULL
);

INSERT INTO vocabulary (level_number, korean, english) VALUES
  (1, '한국',   'Korea'),
  (1, '도시',   'city'),
  (1, '이름',   'name'),
  (1, '저',     'I (formal)'),
  (1, '나',     'I (informal)'),
  (1, '남자',   'man'),
  (1, '여자',   'woman'),
  (1, '이',     'this'),
  (1, '그',     'that'),
  (1, '저',     'that (far away)'),
  (1, '것',     'thing'),
  (1, '이것',   'this thing'),
  (1, '그것',   'that thing'),
  (1, '저것',   'that thing (far away)'),
  (1, '의자',   'chair'),
  (1, '탁자',   'table'),
  (1, '선생님', 'teacher'),
  (1, '침대',   'bed'),
  (1, '집',     'house'),
  (1, '차',     'car'),
  (1, '사람',   'person'),
  (1, '책',     'book'),
  (1, '컴퓨터', 'computer'),
  (1, '나무',   'tree'),
  (1, '소파',   'sofa'),
  (1, '지갑',   'wallet'),
  (1, '방',     'room'),
  (1, '문',     'door'),
  (1, '의사',   'doctor'),
  (1, '학생',   'student');
