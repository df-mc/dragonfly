# Jukebox Translations Design

## Goal

Implement [df-mc/dragonfly#987](https://github.com/df-mc/dragonfly/issues/987) so the popup shown when a player inserts a music disc uses Minecraft's locale-aware translation system instead of the hard-coded English `Now playing:` prefix.

## Scope

The change will use the vanilla `%record.nowPlaying` translation key and preserve Dragonfly's existing `Author - DisplayName` value as its single parameter. Translating individual author and disc names is outside the issue's scope because Minecraft's `record.nowPlaying` translation only localises the surrounding message.

## Design

- Add `chat.MessageNowPlaying`, a one-parameter `chat.Translation` with `%record.nowPlaying` as its client translation identifier and `Now playing: %v` as its server-side fallback.
- Add a translated counterpart to the player's jukebox-popup API. It will accept a `chat.Translation` and parameters, use the player's locale, and delegate packet creation to the session layer, matching `Player.Messaget`.
- Extend session jukebox-popup sending so translated popups use `packet.TextTypeJukeboxPopup`, set `NeedsTranslation`, and populate the message and parameters from Dragonfly's existing translation abstraction.
- Update `block.Jukebox.Activate` and its local user interface to send `chat.MessageNowPlaying` with the existing `Author - DisplayName` text.
- Keep the existing raw `SendJukeboxPopup` API unchanged for compatibility.

## Data Flow

When a player inserts a music disc, `Jukebox.Activate` constructs the existing disc credit text and passes it with `chat.MessageNowPlaying` to the player. The player supplies its locale to the session. The session resolves the translation identifier and parameters and emits a jukebox-popup text packet for client-side localisation.

## Error Handling and Compatibility

The translation follows the established `chat.Translation` contract: an incorrect parameter count panics at the API boundary, as it does for other translated messages. No new I/O or recoverable error path is introduced. The raw popup API remains available and its packet behaviour remains unchanged.

## Testing

Implementation will follow test-driven development:

1. Add a focused session test proving a translated jukebox popup has the jukebox text type, translation flag, vanilla key, and parameter list; run it and confirm it fails before implementation.
2. Add a focused block test proving jukebox activation uses the translated API and supplies the expected disc credit; run it and confirm it fails before implementation.
3. Implement the minimum changes needed to pass those tests, then run affected package tests and the complete repository test suite.

## Delivery

Work is based on the current `df-mc/dragonfly` `master` branch in `agent/issue-987-jukebox-translations`. Changes will remain local; no pull request will be opened.
