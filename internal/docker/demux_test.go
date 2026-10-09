package docker

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"

	"github.com/docker/docker/pkg/stdcopy"
)

func TestStreamDemuxLines(t *testing.T) {
	// Build a multiplexed stream using stdcopy.NewStdWriter
	var rawBuf bytes.Buffer
	stdoutWriter := stdcopy.NewStdWriter(&rawBuf, stdcopy.Stdout)
	stderrWriter := stdcopy.NewStdWriter(&rawBuf, stdcopy.Stderr)

	_, _ = stdoutWriter.Write([]byte("line 1 from stdout\nline 2 from stdout\n"))
	_, _ = stderrWriter.Write([]byte("line 3 from stderr\n"))

	var collected []string
	var mu sync.Mutex

	err := StreamDemuxLines(context.Background(), &rawBuf, func(line string) error {
		mu.Lock()
		defer mu.Unlock()
		collected = append(collected, line)
		return nil
	})

	if err != nil {
		t.Fatalf("StreamDemuxLines failed: %v", err)
	}

	want := []string{
		"line 1 from stdout",
		"line 2 from stdout",
		"line 3 from stderr",
	}

	if len(collected) != len(want) {
		t.Fatalf("collected %d lines, want %d: %v", len(collected), len(want), collected)
	}

	for i := range want {
		if collected[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, collected[i], want[i])
		}
	}
}

func TestStreamDemuxLinesEmpty(t *testing.T) {
	var rawBuf bytes.Buffer
	var count int

	err := StreamDemuxLines(context.Background(), &rawBuf, func(line string) error {
		count++
		return nil
	})

	if err != nil {
		t.Fatalf("StreamDemuxLines on empty input failed: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 lines, got %d", count)
	}
}

func TestStreamDemuxLinesCancellation(t *testing.T) {
	pr, pw := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		stdoutWriter := stdcopy.NewStdWriter(pw, stdcopy.Stdout)
		_, _ = stdoutWriter.Write([]byte("first line\n"))
		// Hold pipe open without writing further
	}()

	var received []string
	err := StreamDemuxLines(ctx, pr, func(line string) error {
		received = append(received, line)
		cancel() // cancel immediately upon first line
		return nil
	})

	_ = pw.Close()

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if len(received) != 1 || received[0] != "first line" {
		t.Errorf("expected 1 line, got %v", received)
	}
}

func TestStreamDemuxLinesLargeLine(t *testing.T) {
	// 256 KiB line - well above default 64 KiB bufio.Scanner limit
	largePayload := make([]byte, 256*1024)
	for i := range largePayload {
		largePayload[i] = 'A' + byte(i%26)
	}

	var rawBuf bytes.Buffer
	stdoutWriter := stdcopy.NewStdWriter(&rawBuf, stdcopy.Stdout)
	_, _ = stdoutWriter.Write(append(largePayload, '\n'))

	var received string
	err := StreamDemuxLines(context.Background(), &rawBuf, func(line string) error {
		received = line
		return nil
	})

	if err != nil {
		t.Fatalf("StreamDemuxLines failed on 256 KiB line: %v", err)
	}

	if len(received) != len(largePayload) {
		t.Fatalf("received length %d, want %d", len(received), len(largePayload))
	}
	if received != string(largePayload) {
		t.Fatalf("received content mismatch on large line")
	}
}

func TestStreamDemuxLinesExceedingMaxBuffer(t *testing.T) {
	// Line exceeding MaxLogBufferSize (2 MiB)
	tooLarge := make([]byte, MaxLogBufferSize+1024)
	for i := range tooLarge {
		tooLarge[i] = 'X'
	}

	var rawBuf bytes.Buffer
	stdoutWriter := stdcopy.NewStdWriter(&rawBuf, stdcopy.Stdout)
	_, _ = stdoutWriter.Write(append(tooLarge, '\n'))

	err := StreamDemuxLines(context.Background(), &rawBuf, func(line string) error {
		return nil
	})

	if err == nil {
		t.Fatalf("expected error when line exceeds MaxLogBufferSize, got nil")
	}
}

func TestStreamDemuxLinesCallbackError(t *testing.T) {
	var rawBuf bytes.Buffer
	stdoutWriter := stdcopy.NewStdWriter(&rawBuf, stdcopy.Stdout)
	_, _ = stdoutWriter.Write([]byte("line 1\nline 2\n"))

	expectedErr := io.ErrUnexpectedEOF
	err := StreamDemuxLines(context.Background(), &rawBuf, func(line string) error {
		return expectedErr
	})

	if err != expectedErr {
		t.Fatalf("expected callback error %v, got %v", expectedErr, err)
	}
}
