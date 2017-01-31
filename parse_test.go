package dockerfile

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseReaderParseError(t *testing.T) {
	dockerfile := "FROM ubuntu:xenial\nCMD [\"echo\", 1]"
	_, err := ParseReader(bytes.NewBufferString(dockerfile))
	assert.IsType(t, ParseError{}, err)
}

func TestParseReader(t *testing.T) {
	dockerfile := `FROM ubuntu:xenial
RUN echo hi > /etc/hi.conf
CMD ["echo"]
ONBUILD ADD foo bar
ONBUILD RUN ["cat", "bar"]
`
	cmds, err := ParseReader(bytes.NewBufferString(dockerfile))
	assert.Nil(t, err)
	expected := []Command{
		Command{
			Cmd:       "from",
			Original:  "FROM ubuntu:xenial",
			StartLine: 1,
			Value:     []string{"ubuntu:xenial"},
		},
		Command{
			Cmd:       "run",
			Original:  "RUN echo hi > /etc/hi.conf",
			StartLine: 2,
			Value:     []string{"echo hi > /etc/hi.conf"},
		},
		Command{
			Cmd:       "cmd",
			Json:      true,
			Original:  "CMD [\"echo\"]",
			StartLine: 3,
			Value:     []string{"echo"},
		},
		Command{
			Cmd:       "onbuild",
			SubCmd:    "add",
			Original:  "ONBUILD ADD foo bar",
			StartLine: 4,
			Value:     []string{"foo", "bar"},
		},
		Command{
			Cmd:       "onbuild",
			SubCmd:    "run",
			Json:      true,
			Original:  "ONBUILD RUN [\"cat\", \"bar\"]",
			StartLine: 5,
			Value:     []string{"cat", "bar"},
		},
	}
	assert.Equal(t, expected, cmds)
}

func TestParseFileIOError(t *testing.T) {
	_, err := ParseFile("Dockerfile.dne")
	assert.IsType(t, IOError{}, err)
	assert.Regexp(t, "^.*Dockerfile.dne.*$", err.Error())
}

func TestParseFile(t *testing.T) {
	cmds, err := ParseFile("testfiles/Dockerfile.ok")
	assert.Nil(t, err)
	expected := []Command{
		Command{
			Cmd:       "from",
			Original:  "FROM ubuntu:xenial",
			StartLine: 1,
			Value:     []string{"ubuntu:xenial"},
		},
		Command{
			Cmd:       "cmd",
			Original:  "CMD [\"echo\", \"hi\"]",
			StartLine: 2,
			Json:      true,
			Value:     []string{"echo", "hi"},
		},
	}
	assert.Equal(t, expected, cmds)
}
