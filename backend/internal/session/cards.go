package session

type rawCard struct {
	ID      int
	Korean  string
	English string
}

func HardCodedCards() []rawCard {
	return []rawCard{
		{ID: 1, Korean: "저는 학생입니다", English: "I am a student"},
		{ID: 2, Korean: "그녀는 의사입니다", English: "She is a doctor"},
		{ID: 3, Korean: "물을 마시고 싶어요", English: "I want to drink water"},
		{ID: 4, Korean: "이 책이 얼마예요?", English: "How much is this book?"},
		{ID: 5, Korean: "어디에서 왔어요?", English: "Where are you from?"},
		{ID: 6, Korean: "내일 만나요", English: "Let's meet tomorrow"},
		{ID: 7, Korean: "한국 음식을 좋아해요", English: "I like Korean food"},
		{ID: 8, Korean: "이름이 뭐예요?", English: "What is your name?"},
		{ID: 9, Korean: "버스를 타고 가요", English: "I go by bus"},
		{ID: 10, Korean: "날씨가 좋아요", English: "The weather is nice"},
	}
}
