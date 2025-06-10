package llm

var OPENAI_GPT4o = Model{
	Name:        "gpt-4o",
	ContextSize: 128000, // 128k tokens
}

var OPENAI_GPT41_Mini = Model{
	Name:        "gpt-4.1-mini",
	ContextSize: 1000000, // 1M tokens
}
