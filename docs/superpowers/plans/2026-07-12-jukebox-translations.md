# Jukebox Translations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Localise the jukebox now-playing popup with Minecraft's `%record.nowPlaying` translation.

**Architecture:** Reuse `chat.Translation` from the block through `player.Player` into `session.Session`. Preserve the raw popup API, while the translated path produces a jukebox-popup `packet.Text` carrying the vanilla translation key and its disc-credit parameter.

**Tech Stack:** Go 1.26, Dragonfly player/chat/session APIs, gophertunnel Bedrock packets, Go `testing`.

## Global Constraints

- Base all work on `df-mc/dragonfly` `master`.
- Preserve `SendJukeboxPopup` compatibility.
- Keep `Author - DisplayName` as the single untranslated parameter.
- Do not open a pull request.

---

### Task 1: Translated Jukebox Popup Pipeline

**Files:**
- Modify: `server/player/chat/translate.go`
- Modify: `server/session/text.go`
- Create: `server/session/text_test.go`
- Modify: `server/player/player.go`

**Interfaces:**
- Produces: `chat.MessageNowPlaying chat.Translation`
- Produces: `(*session.Session).SendJukeboxTranslation(t chat.Translation, l language.Tag, a []any)`
- Produces: `(*player.Player).SendJukeboxPopupt(t chat.Translation, a ...any)`

- [ ] **Step 1: Write the failing session test**

Create a session with buffered `packets` and `closeBackground` channels, call `SendJukeboxTranslation(chat.MessageNowPlaying, language.French, []any{"C418 - cat"})`, receive the `*packet.Text`, and assert `TextTypeJukeboxPopup`, `NeedsTranslation == true`, `Message == "%record.nowPlaying"`, and `Parameters == []string{"C418 - cat"}`.

- [ ] **Step 2: Run the test to verify RED**

Run: `go test ./server/session -run TestSendJukeboxTranslation -count=1`

Expected: build failure because `SendJukeboxTranslation` and `chat.MessageNowPlaying` do not exist.

- [ ] **Step 3: Implement the translation pipeline**

Add:

```go
var MessageNowPlaying = Translate(str("%record.nowPlaying"), 1, `Now playing: %v`)
```

Implement `Session.SendJukeboxTranslation` like `SendTranslation`, but with `packet.TextTypeJukeboxPopup`. Add `Player.SendJukeboxPopupt`, delegating with the player's locale.

- [ ] **Step 4: Run the focused test to verify GREEN**

Run: `go test ./server/session -run TestSendJukeboxTranslation -count=1`

Expected: PASS.

- [ ] **Step 5: Commit the pipeline**

```bash
git add server/player/chat/translate.go server/session/text.go server/session/text_test.go server/player/player.go
git commit -m "player: add translated jukebox popups"
```

### Task 2: Use the Translation When Activating Jukeboxes

**Files:**
- Modify: `server/block/jukebox.go`
- Create: `server/block/jukebox_test.go`

**Interfaces:**
- Consumes: `chat.MessageNowPlaying`
- Consumes: `SendJukeboxPopupt(chat.Translation, ...any)`
- Produces: locale-aware now-playing behaviour from `Jukebox.Activate`

- [ ] **Step 1: Write the failing activation test**

Create a minimal `item.User` recorder holding `item.NewStack(item.MusicDisc{DiscType: sound.DiscCat()}, 1)` and recording the translation and arguments passed to `SendJukeboxPopupt`. Activate an empty jukebox in a synchronous world transaction. Assert the translation is `chat.MessageNowPlaying`, the argument is `C418 - cat`, and the use context subtracts one item.

- [ ] **Step 2: Run the test to verify RED**

Run: `go test ./server/block -run TestJukeboxActivateSendsTranslatedPopup -count=1`

Expected: FAIL because `Jukebox.Activate` still calls the raw popup method.

- [ ] **Step 3: Implement the translated activation**

Change `jukeboxUser` to require `SendJukeboxPopupt(chat.Translation, ...any)`, import `server/player/chat`, and call:

```go
u.SendJukeboxPopupt(chat.MessageNowPlaying, fmt.Sprintf("%v - %v", m.DiscType.Author(), m.DiscType.DisplayName()))
```

- [ ] **Step 4: Run focused and affected tests to verify GREEN**

Run: `go test ./server/block ./server/session ./server/player/... -count=1`

Expected: PASS.

- [ ] **Step 5: Commit the behaviour**

```bash
git add server/block/jukebox.go server/block/jukebox_test.go
git commit -m "block: localise jukebox now-playing popup"
```

### Task 3: Repository Verification

**Files:**
- Verify only; no planned modifications.

**Interfaces:**
- Consumes: completed Tasks 1 and 2.
- Produces: a clean, tested local branch ready for later review.

- [ ] **Step 1: Format and inspect changes**

Run: `gofmt -w server/player/chat/translate.go server/session/text.go server/session/text_test.go server/player/player.go server/block/jukebox.go server/block/jukebox_test.go`

Run: `git diff --check && git status --short`

Expected: no whitespace errors and no unexpected files.

- [ ] **Step 2: Run the full suite**

Run: `go test ./...`

Expected: PASS.

- [ ] **Step 3: Review branch scope**

Run: `git log --oneline upstream/master..HEAD && git diff --stat upstream/master...HEAD`

Expected: only the design/plan and jukebox translation commits/files are present; no PR is opened.
