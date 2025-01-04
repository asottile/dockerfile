package dockerfile

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllCmds(t *testing.T) {
	ret := AllCmds()
	assert.Equal(t, ret[:3], []string{"add", "arg", "cmd"})
}

func TestParseReaderParseError(t *testing.T) {
	dockerfile := "FROM ubuntu:xenial\nCMD [\"echo\", 1]"
	_, err := ParseReader(bytes.NewBufferString(dockerfile))
	assert.IsType(t, ParseError{}, err)
}

func TestParseReader(t *testing.T) {
	dockerfile := `FROM ubuntu:xenial
RUN echo hi > /etc/hi.conf
CMD ["echo"]
HEALTHCHECK --retries=5 CMD echo hi
ONBUILD ADD foo bar
ONBUILD RUN ["cat", "bar"]
`
	cmds, err := ParseReader(bytes.NewBufferString(dockerfile))
	assert.Nil(t, err)
	expected := []Command{
		Command{
			Cmd:       "FROM",
			Original:  "FROM ubuntu:xenial",
			StartLine: 1,
			EndLine:   1,
			Flags:     []string{},
			Value:     []string{"ubuntu:xenial"},
		},
		Command{
			Cmd:       "RUN",
			Original:  "RUN echo hi > /etc/hi.conf",
			StartLine: 2,
			EndLine:   2,
			Flags:     []string{},
			Value:     []string{"echo hi > /etc/hi.conf"},
		},
		Command{
			Cmd:       "CMD",
			Json:      true,
			Original:  "CMD [\"echo\"]",
			StartLine: 3,
			EndLine:   3,
			Flags:     []string{},
			Value:     []string{"echo"},
		},
		Command{
			Cmd:       "HEALTHCHECK",
			SubCmd:    "",
			Original:  "HEALTHCHECK --retries=5 CMD echo hi",
			StartLine: 4,
			EndLine:   4,
			Flags:     []string{"--retries=5"},
			Value:     []string{"CMD", "echo hi"},
		},
		Command{
			Cmd:       "ONBUILD",
			SubCmd:    "ADD",
			Original:  "ONBUILD ADD foo bar",
			StartLine: 5,
			EndLine:   5,
			Flags:     []string{},
			Value:     []string{"foo", "bar"},
		},
		Command{
			Cmd:       "ONBUILD",
			SubCmd:    "RUN",
			Json:      true,
			Original:  "ONBUILD RUN [\"cat\", \"bar\"]",
			StartLine: 6,
			EndLine:   6,
			Flags:     []string{},
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
			Cmd:       "FROM",
			Original:  "FROM ubuntu:xenial",
			StartLine: 1,
			EndLine:   1,
			Flags:     []string{},
			Value:     []string{"ubuntu:xenial"},
		},
		Command{
			Cmd:       "CMD",
			Original:  "CMD [\"echo\", \"hi\"]",
			StartLine: 2,
			EndLine:   2,
			Json:      true,
			Flags:     []string{},
			Value:     []string{"echo", "hi"},
		},
	}
	assert.Equal(t, expected, cmds)
}

func TestParseReaderHeredocs(t *testing.T) {
	dockerfile := `RUN 3<<EOF
source $HOME/.bashrc && echo $HOME
echo "Hello" >> /hello
echo "World!" >> /hello
EOF
`
	cmds, err := ParseReader(bytes.NewBufferString(dockerfile))
	assert.Nil(t, err)
	expected := []Command{
		Command{
			Cmd:       "RUN",
			Original:  dockerfile,
			StartLine: 1,
			EndLine:   5,
			Flags:     []string{},
			Value:     []string{"3<<EOF"},
			Heredocs: []Heredoc{
				Heredoc{
					Name:           "EOF",
					FileDescriptor: 3,
					Content:        "source $HOME/.bashrc && echo $HOME\necho \"Hello\" >> /hello\necho \"World!\" >> /hello\n"},
			},
		},
	}
	assert.Equal(t, expected, cmds)
}

func TestParseReaderHeredocsMultiple(t *testing.T) {
	dockerfile := `COPY <<FILE1 <<FILE2 /dest
content 1
FILE1
content 2
FILE2
`
	cmds, err := ParseReader(bytes.NewBufferString(dockerfile))
	assert.Nil(t, err)
	expected := []Command{
		Command{
			Cmd:       "COPY",
			Original:  dockerfile,
			StartLine: 1,
			EndLine:   5,
			Flags:     []string{},
			Value:     []string{"<<FILE1", "<<FILE2", "/dest"},
			Heredocs: []Heredoc{
				Heredoc{
					Name:           "FILE1",
					FileDescriptor: 0,
					Content:        "content 1\n"},
				Heredoc{
					Name:           "FILE2",
					FileDescriptor: 0,
					Content:        "content 2\n"},
			},
		},
	}
	assert.Equal(t, expected, cmds)
}
