package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dgr8akki/nano-gameplays/internal/analysis"
	"github.com/dgr8akki/nano-gameplays/internal/handoff"
)

var allowedVideoExtensions = map[string]struct{}{
	"mp4":  {},
	"mov":  {},
	"mkv":  {},
	"webm": {},
}

// runIdentifyOnce executes the non-interactive identify flow per
// contracts/cli.md. Exit codes: 0 payload emitted, 2 precondition failure.
func runIdentifyOnce(videoPath, modelID string, asJSON bool) int {
	abs, err := preflightVideo(videoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "identify: %v\n", err)
		return 2
	}
	video := buildVideo(abs)

	frames, extractErr := analysis.ExtractFrames(context.Background(), abs)
	if extractErr != nil && len(frames) == 0 {
		// Without frames there is nothing to send; still emit the payload so
		// callers can read analysis.error per the contract.
		emitFailurePayload(video, modelID, extractErr.Error(), asJSON)
		return 0
	}

	meta, rec, identifyErr := analysis.Identify(context.Background(), frames, modelID)
	payload := handoff.HandoffPayload{
		SchemaVersion: handoff.SchemaVersion,
		Video:         video,
		Metadata:      meta,
		Analysis:      rec,
	}
	if identifyErr != nil {
		// Metadata.GameTitle is empty per spec; record already carries the error.
	}
	return emitPayload(payload, asJSON)
}

func preflightVideo(path string) (string, error) {
	if path == "" {
		return "", errors.New("--video PATH is required")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("video missing or unreadable: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory, not a video file", abs)
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(abs), "."))
	if _, ok := allowedVideoExtensions[ext]; !ok {
		return "", fmt.Errorf("unsupported extension %q (want mp4|mov|mkv|webm)", ext)
	}
	return abs, nil
}

func buildVideo(abs string) handoff.GameplayVideo {
	out := handoff.GameplayVideo{Path: abs}
	if info, err := os.Stat(abs); err == nil {
		out.SizeBytes = info.Size()
	}
	out.Extension = strings.ToLower(strings.TrimPrefix(filepath.Ext(abs), "."))
	return out
}

func emitFailurePayload(video handoff.GameplayVideo, modelID, errMsg string, asJSON bool) {
	payload := handoff.HandoffPayload{
		SchemaVersion: handoff.SchemaVersion,
		Video:         video,
		Metadata:      handoff.GameMetadata{Confidence: "low"},
		Analysis: handoff.AnalysisRecord{
			ModelID: modelID,
			Error:   errMsg,
		},
	}
	emitPayload(payload, asJSON)
}

func emitPayload(payload handoff.HandoffPayload, asJSON bool) int {
	if asJSON {
		data, err := json.Marshal(payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "encode payload: %v\n", err)
			return 2
		}
		os.Stdout.Write(data)
		os.Stdout.WriteString("\n")
		return 0
	}
	fmt.Printf("video: %s\n", payload.Video.Path)
	fmt.Printf("game: %s\n", payload.Metadata.GameTitle)
	if payload.Metadata.SceneOrLevelOrMode != "" {
		fmt.Printf("scene: %s\n", payload.Metadata.SceneOrLevelOrMode)
	}
	fmt.Printf("confidence: %s\n", payload.Metadata.Confidence)
	if payload.Analysis.Error != "" {
		fmt.Printf("error: %s\n", payload.Analysis.Error)
	}
	return 0
}
