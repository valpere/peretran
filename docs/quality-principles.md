# Quality Translation Principles

This document describes the quality-focused approaches implemented in peretran, inspired by TranslateBooksWithLLMs and book-translator.

---

## Core Principles

### 1. Two-Pass Translation

The most important quality factor is separating translation from refinement:

**Stage 1: Primary Translation**
- Translates text from source to target language
- Focus on accuracy and meaning preservation
- Uses "professional translator" prompt
- Includes context from previous chunks

**Stage 2: Refinement**
- Acts as "elite literary editor"
- Reviews draft against original
- Improves: flow, idioms, word choice, rhythm
- Keeps draft unchanged if already good

This separation is critical because:
- First pass focuses on getting meaning right
- Second pass focuses on making it read beautifully
- Same LLM can do both with different prompts

---

### 2. Adaptive Context Management

Dynamic context sizing prevents translation failures:

- **Start small**: 2048 tokens (standard models) or higher for "thinking" models
- **Detect truncation**: Check if response was truncated
- **Increase context**: Retry with larger context if needed
- **Track usage**: Learn optimal context size over time

**Handling Overflow:**
- Reduce chunk size by 60%
- Split at sentence boundaries
- Retry with smaller chunk
- Max 3 reduction attempts

---

### 3. Structure Preservation

For structured content (HTML, EPUB, Markdown):

1. **Before translation**: Replace tags with placeholders
   - `<p><span>Hello</span></p>` → `[id0]Hello[id1]`
   - Groups adjacent tags to reduce token count

2. **After translation**: Validate and restore
   - Check expected placeholder count
   - Detect missing/duplicate/mutated placeholders
   - Use correction prompt if needed

---

### 4. Context Continuity

Maintains consistency across chunks:

- **Previous translation**: Last 25 words passed to next chunk
- **Context before/after**: Surrounding text for paragraph flow
- **Context hash**: Included in cache key for differentiation

---

### 5. Engineering Rigor

**Checkpoints**
- Save after each chunk
- Store: translated text, context, progress
- Resume from interruption

**Validation**
- Check translation is in target language
- Reject if source language detected
- Retry with backoff on failure

**Post-Processing Cleanup**
- Remove thinking tags (`<thinking>`, `<think>`)
- Remove instruction echoes ("TEXT TO TRANSLATE:")
- Remove context leaks
- Remove repetition from previous chunks

---

## Additional Features

### Translation Caching

- SQLite-based persistent cache
- Context-aware hashing (includes previous chunk)
- Separate cache for stage 1 and stage 2
- Tracks usage count and last_used

### Terminology Management

- Extract proper nouns from text
- Maintain glossary for consistency
- Include in prompts

### Retry with Backoff

- Up to 3 retry attempts
- Exponential backoff delay
- Return last attempt even if validation fails

---

## Implementation Checklist

For every translation:

- [ ] Translation is in target language
- [ ] No thinking tags in output
- [ ] No instruction echoes
- [ ] No repeated content
- [ ] Formatting preserved
- [ ] Validation passed

---

## References

- TranslateBooksWithLLMs: https://github.com/.../TranslateBooksWithLLMs
- book-translator: https://github.com/.../book-translator
