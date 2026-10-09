package docker

import (
	"bufio"
	"context"
	"io"

	"github.com/docker/docker/pkg/stdcopy"
)

// StreamDemuxLines demultiplexes a Docker container log stream (stdcopy format)
// and invokes onLine for each emitted text line. It stops when the input stream
// reaches EOF, an error occurs, or the context is cancelled.
func StreamDemuxLines(ctx context.Context, src io.Reader, onLine func(line string) error) error {
	pr, pw := io.Pipe()

	errCh := make(chan error, 1)
	go func() {
		defer pw.Close()
		// StdCopy demuxes Docker's 8-byte frame header into stdout and stderr writers.
		// Both stdout and stderr are directed to the same pipe writer.
		_, err := stdcopy.StdCopy(pw, pw, src)
		errCh <- err
	}()

	scanner := bufio.NewScanner(pr)
	scanDone := make(chan error, 1)

	go func() {
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				scanDone <- ctx.Err()
				return
			default:
			}

			if err := onLine(scanner.Text()); err != nil {
				scanDone <- err
				return
			}
		}
		scanDone <- scanner.Err()
	}()

	select {
	case <-ctx.Done():
		pr.Close()
		return ctx.Err()
	case err := <-scanDone:
		pr.Close()
		copyErr := <-errCh
		if err != nil {
			return err
		}
		if copyErr != nil && copyErr != io.EOF {
			return copyErr
		}
		return nil
	}
}
