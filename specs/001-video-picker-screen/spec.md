# Feature Specification: Initial UI — Video Picker & Game Identification

**Feature Branch**: `001-video-picker-screen`  
**Created**: 2026-05-01  
**Status**: Draft  
**Input**: User description: "initial ui screen - It will be split screen into 4 part header, footer, middle left, middle right. In Header, we will show the name of the app, in footer, we will show the keys and their meaning. In middle right screen, it will show a cool ps5 ascii animation till the user is selecting a gamevideo. In middle left, it should be  able to pick the video with the help of keys, it will show directory structure as well. once the file is selected, it will show a \"Analyzing...\" text with a cool animation on the middle right screen. internally it will take a screenshot or may be 2 and then send it to claude code and get the game info pre populated like game name, scene etc on the right middle screen. User can also edit it before proceeding it to the next screen."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Pick a gameplay video from disk (Priority: P1)

A user opens the app, sees a split-screen interface, and uses keyboard
navigation in the left pane to browse their filesystem and select a
PlayStation gameplay video file. Once selected, they can confirm and proceed
toward the next screen. This is the core entry point of the entire product;
without it, no other capability is reachable.

**Why this priority**: This is the application's front door. Every downstream
feature (analysis, shorts generation, publishing) depends on a user
successfully identifying and selecting a source video. If only this story
ships, the user has a working file picker that can hand off to later screens.

**Independent Test**: Launch the app in a directory containing at least one
video file. Navigate using only the keyboard, select a video, and confirm
that the app records the selected file path and signals readiness to advance.
No AI, no analysis, no animations are required for this slice to deliver
value.

**Acceptance Scenarios**:

1. **Given** the app is launched in a directory tree containing video files,
   **When** the user navigates to and selects a video file using keyboard
   keys shown in the footer, **Then** the app records the selected file and
   indicates the next step is ready.
2. **Given** the user is browsing a folder, **When** they enter a
   subdirectory or go up one level using the documented keys, **Then** the
   left pane updates to reflect the new directory contents without losing
   focus.
3. **Given** a folder contains both video and non-video files, **When** the
   user views that folder, **Then** non-video files are clearly
   distinguishable from selectable video files.
4. **Given** the user attempts to select an unsupported file type, **When**
   they confirm the selection, **Then** the app rejects the choice with a
   clear, non-blocking message and keeps the picker active.

---

### User Story 2 - Auto-fill game metadata from the selected video (Priority: P2)

After selecting a gameplay video, the user sees the right pane transition to
an "Analyzing…" state. The app extracts one or two representative frames
from the video and uses Claude to identify game-level metadata (game title,
in-game scene/level, and other notable context). The identified fields
appear in the right pane, editable by the user, before they proceed.

**Why this priority**: This is the differentiating intelligence of the
product — it removes manual data entry from the user's path. It depends on
P1 (a selected video is required) but is not required for P1 to deliver
value, so it can ship in a second iteration.

**Independent Test**: With a video already selected (P1), trigger the
analysis step and verify (a) the right pane shows an analyzing state, (b)
identified metadata fields appear within the success-criteria time budget,
(c) every field is editable via keyboard, and (d) the user can confirm and
advance.

**Acceptance Scenarios**:

1. **Given** a video has been selected, **When** the analysis step runs,
   **Then** the right pane shows an "Analyzing…" indicator with visible
   progress until results arrive.
2. **Given** analysis succeeds, **When** results are returned, **Then** the
   right pane displays at minimum: detected game title, detected scene /
   level / mode, and a confidence indicator, all in editable fields.
3. **Given** detected fields are shown, **When** the user edits a field
   using the keyboard, **Then** the edit is reflected immediately and
   persists when they proceed.
4. **Given** analysis fails or returns nothing usable, **When** the failure
   is surfaced, **Then** the user sees a clear message and is given empty
   editable fields so they can proceed by entering metadata manually.
5. **Given** analysis is in progress, **When** the user cancels or selects a
   different video, **Then** the in-flight analysis is abandoned and the
   right pane resets cleanly.

---

### User Story 3 - Polished split-screen shell (Priority: P3)

The user experiences the full intended visual shell at all times: a header
showing the application name, a footer showing the currently-applicable key
bindings with their meanings, a left pane occupied by the file picker, and a
right pane that shows a PS5-themed ASCII animation while the user is still
choosing a video and a different "analyzing" animation while metadata is
being fetched.

**Why this priority**: The shell is what makes the product feel finished.
It depends on P1 and P2 being functional, but the underlying flow can be
demonstrated without animations or the styled shell. It is therefore the
last slice to ship.

**Independent Test**: Launch the app and, without performing any work,
visually confirm that all four regions (header, footer, left, right) are
present, that the right pane is animating, and that the footer key list
matches the keys that actually work in the current focus context.

**Acceptance Scenarios**:

1. **Given** the app is launched, **When** the user sees the initial screen,
   **Then** the header shows the application name, the footer lists the
   currently-applicable keys with short descriptions, and both middle panes
   are visible side-by-side.
2. **Given** the user has not yet selected a video, **When** they look at
   the right pane, **Then** they see a PS5-themed ASCII animation that
   continues to play smoothly without distracting from the picker.
3. **Given** a video has been selected, **When** analysis begins, **Then**
   the right pane swaps from the idle animation to an analyzing animation
   accompanied by an "Analyzing…" label.
4. **Given** the user moves between focusable regions, **When** focus
   changes, **Then** the footer updates to show the keys relevant to the
   newly-focused region.

---

### Edge Cases

- The starting directory contains no video files at any depth.
- The starting directory contains thousands of entries (rendering and
  scrolling must remain responsive).
- The user lacks permission to read a directory they navigate into.
- The user selects a video file that is corrupt or zero-byte, so frame
  extraction cannot produce a usable image.
- Frame extraction succeeds but the resulting image is uninformative
  (loading screen, all-black cutscene, menu).
- The Claude request times out, errors, or returns content that does not
  match the expected metadata shape.
- The user resizes the terminal (smaller than the layout's minimum, or
  during analysis).
- The user selects a new video while a previous analysis is still running.
- The user has restricted or no internet connectivity at the moment of
  analysis.
- A symbolic link points to a directory outside the apparent tree (cycle
  risk).
- The user navigates above the starting directory and reaches the
  filesystem root.

## Requirements *(mandatory)*

### Functional Requirements

**Layout & Shell**

- **FR-001**: The application MUST present, on launch, a single screen
  divided into four regions: header (top), footer (bottom), and a middle
  area split into a left pane and a right pane.
- **FR-002**: The header MUST display the application's name.
- **FR-003**: The footer MUST display the keys currently available to the
  user along with a short description of what each key does, and MUST
  update when the active context changes.
- **FR-004**: While the user has not yet selected a video, the right pane
  MUST display a PS5-themed ASCII animation.
- **FR-005**: After a video has been selected and while metadata is being
  fetched, the right pane MUST display an "Analyzing…" label paired with
  a distinct animation.

**File Selection (Left Pane)**

- **FR-006**: The left pane MUST allow the user to browse the filesystem
  starting from a defined root directory.
- **FR-007**: The left pane MUST visualize the directory structure (folders
  and their contents) and the user's current location within it.
- **FR-008**: The user MUST be able to navigate the directory structure
  using only the keyboard (move up/down within a folder, enter a folder,
  return to the parent folder).
- **FR-009**: The user MUST be able to confirm a selection of a video file
  using a single, documented key.
- **FR-010**: The left pane MUST visually distinguish video files from
  files that cannot be selected.
- **FR-011**: When the user attempts to confirm a selection on a non-video
  file, the application MUST refuse the selection without leaving the
  picker and MUST inform the user briefly.

**Analysis & Metadata (Right Pane)**

- **FR-012**: When a video is selected, the application MUST capture one
  or two representative frames from that video for analysis.
- **FR-013**: The application MUST submit the captured frame(s) to Claude
  to obtain pre-filled metadata about the gameplay shown.
- **FR-014**: The pre-filled metadata MUST include at minimum: detected
  game title and detected in-game scene/level/mode.
- **FR-015**: The application MUST display the returned metadata in the
  right pane in editable fields.
- **FR-016**: The user MUST be able to edit each metadata field using the
  keyboard before proceeding.
- **FR-017**: The user MUST be able to confirm the (possibly edited)
  metadata with a documented key in order to advance to the next screen.
- **FR-018**: If Claude returns no usable result or fails entirely, the
  application MUST surface that fact to the user, present empty editable
  metadata fields, and still allow the user to proceed by entering
  metadata manually.
- **FR-019**: If frame extraction fails, the application MUST inform the
  user and still allow manual entry of metadata to proceed.
- **FR-020**: If the user selects a different video while an analysis is
  in flight, the application MUST abandon the in-flight analysis and
  reset the right pane state for the new selection.

**Feedback & Resilience**

- **FR-021**: All long-running activities (analysis, frame extraction,
  Claude call) MUST surface progress and never appear frozen.
- **FR-022**: All user-facing failure messages MUST be actionable: they
  state what failed and what the user can do next.

### Key Entities *(include if feature involves data)*

- **Gameplay Video**: A user-supplied video file representing a session of
  PlayStation gameplay. Identified by its filesystem path. The unit of work
  selected on this screen and passed to later screens.
- **Captured Frame**: One or two still images extracted from the gameplay
  video, used solely as input to game-identification analysis.
- **Game Metadata**: A small set of fields describing the gameplay shown,
  initially populated by analysis and editable by the user. At minimum:
  detected game title, detected scene/level/mode, and a confidence
  indicator. Carried forward to the next screen alongside the gameplay
  video selection.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: From app launch to confirmed metadata, a user with a known
  video in a typical local directory can complete the flow in under 30
  seconds when analysis succeeds on the first attempt.
- **SC-002**: For at least 80% of mainstream PS5 titles in user testing,
  the auto-detected game title is correct and requires no edit.
- **SC-003**: At least 90% of users complete video selection in their
  first session without referring to external documentation; the footer
  key hints are sufficient.
- **SC-004**: 100% of analysis failure cases (network failure, model
  error, unusable frame) are recoverable on this screen via manual entry,
  with no need to restart the app.
- **SC-005**: For directories of up to 1,000 entries, navigation between
  items feels immediate to the user (no perceptible lag during keypress
  movement).
- **SC-006**: When analysis succeeds, the user sees populated metadata
  fields within 5 seconds of confirming their video selection on a
  reference internet connection.
- **SC-007**: The shell renders correctly in terminal sizes of at least
  100 columns by 30 rows; below that, the application gives the user a
  clear message instead of a broken layout.

## Assumptions

- The starting directory for the file picker is the directory from which
  the application is launched. The user can navigate up to a parent
  directory within the picker.
- Supported video file types are the common gameplay capture formats (for
  example, MP4 and MOV). Other media types are visible during browsing but
  not selectable.
- The application has working network access at the time of analysis. The
  product handles unavailability gracefully (per FR-018) but does not
  attempt fully offline detection.
- The user is operating an interactive terminal at least 100×30 in size and
  has a keyboard. Mouse input is out of scope for this screen.
- Authentication and quota for the underlying Claude integration are
  handled at the application level and are not part of this screen's UX.
- This screen produces a "selected video + metadata" handoff to a later
  screen, but the design of that next screen is out of scope here.
- Captured frames are used only as analysis input and are not retained as
  user-visible artifacts on this screen.
