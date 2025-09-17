package shared

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFollow(t *testing.T) {
	tests := []struct {
		name           string
		log            string
		wantStdoutFile string
		wantStderrFile string
	}{
		{
			name:           "sample log 1",
			log:            "testdata/log-1-input.txt",
			wantStdoutFile: "testdata/log-1-want.txt",
		},
		{
			name:           "sample log 2",
			log:            "testdata/log-2-input.txt",
			wantStdoutFile: "testdata/log-2-want.txt",
		},
		{
			name:           "sample log 3 (tolerant parse failures)",
			log:            "testdata/log-3-synthetic-failures-input.txt",
			wantStdoutFile: "testdata/log-3-synthetic-failures-want.txt",
			wantStderrFile: "testdata/log-3-synthetic-failures-want-stderr.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := os.ReadFile(tt.log)
			require.NoError(t, err)

			// Normalize CRLF to LF to make the tests OS-agnostic.
			raw = []byte(strings.ReplaceAll(string(raw), "\r\n", "\n"))

			lines := slices.DeleteFunc(strings.Split(string(raw), "\n"), func(line string) bool {
				return line == ""
			})

			var hits int
			fetcher := func() ([]byte, error) {
				hits++
				if hits > len(lines) {
					require.FailNow(t, "too many API calls")
				}
				return []byte(strings.Join(lines[0:hits], "\n\n")), nil
			}

			ios, _, stdout, stderr := iostreams.Test()

			err = NewLogRenderer().Follow(fetcher, stdout, ios)
			require.NoError(t, err)

			// Handy note for updating the testdata files when they change:
			// ext := filepath.Ext(tt.log)
			// stripped := strings.TrimSuffix(tt.log, ext)
			// stripped = strings.TrimSuffix(stripped, "-input")
			// os.WriteFile(stripped+"-want"+ext, stdout.Bytes(), 0644)
			// if tt.wantStderrFile != "" {
			// 	os.WriteFile(stripped+"-want-stderr"+ext, stderr.Bytes(), 0644)
			// }

			wantStdout, err := os.ReadFile(tt.wantStdoutFile)
			require.NoError(t, err)

			// Normalize CRLF to LF to make the tests OS-agnostic.
			wantStdout = []byte(strings.ReplaceAll(string(wantStdout), "\r\n", "\n"))

			assert.Equal(t, string(wantStdout), stdout.String())

			if tt.wantStderrFile != "" {
				wantStderr, err := os.ReadFile(tt.wantStderrFile)
				require.NoError(t, err)

				// Normalize CRLF to LF to make the tests OS-agnostic.
				wantStderr = []byte(strings.ReplaceAll(string(wantStderr), "\r\n", "\n"))

				assert.Equal(t, string(wantStderr), stderr.String())
			} else {
				require.Empty(t, stderr, "expected no stderr output")
			}
		})
	}
}
