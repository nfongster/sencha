# Future Ideas

## Per-category vocab sampling

Add a `category` property to vocabulary entries: noun, verb, adjective, adverb, etc.

Then change the sampling rule from "sample 50 random words across the entire vocab set (or all if <= 50)" to "sample N random words from each category (or all words for that category if <= N)". This ensures an even distribution of word types in every session rather than letting the draw be lopsided.
