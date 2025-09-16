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
		name string
		log  string
		want string
	}{
		{
			name: "sample log 1",
			log:  "testdata/log-1-input.txt",
			want: "testdata/log-1-want.txt",
		},
		{
			name: "sample log 2",
			log:  "testdata/log-2-input.txt",
			want: "testdata/log-2-want.txt",
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

			ios, _, stdout, _ := iostreams.Test()

			err = NewLogRenderer().Follow(fetcher, stdout, ios)
			require.NoError(t, err)

			// Handy note for updating the testdata files when they change:
			// ext := filepath.Ext(tt.log)
			// stripped := strings.TrimSuffix(tt.log, ext)
			// os.WriteFile(stripped+".want"+ext, stdout.Bytes(), 0644)

			want, err := os.ReadFile(tt.want)
			require.NoError(t, err)

			// Normalize CRLF to LF to make the tests OS-agnostic.
			want = []byte(strings.ReplaceAll(string(want), "\r\n", "\n"))

			assert.Equal(t, string(want), stdout.String())
		})
	}
}
