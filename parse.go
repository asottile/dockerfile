package dockerfile

import (
	"io"
	"os"
	"sort"

	"github.com/moby/buildkit/frontend/dockerfile/command"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

// Represents info about a heredoc.
type Heredoc struct {
	Name           string
	FileDescriptor uint
	Content        string
}

// Represents a single line (layer) in a Dockerfile.
// For example `FROM ubuntu:xenial`
type Command struct {
	Cmd       string    // lowercased command name (ex: `from`)
	SubCmd    string    // for ONBUILD only this holds the sub-command
	Json      bool      // whether the value is written in json form
	Original  string    // The original source line
	StartLine int       // The original source line number which starts this command
	EndLine   int       // The original source line number which ends this command
	Flags     []string  // Any flags such as `--from=...` for `COPY`.
	Value     []string  // The contents of the command (ex: `ubuntu:xenial`)
	Heredocs  []Heredoc // Extra heredoc content attachments
}

// A failure in opening a file for reading.
type IOError struct {
	Msg string
}

func (e IOError) Error() string {
	return e.Msg
}

// A failure in parsing the file as a dockerfile.
type ParseError struct {
	Msg string
}

func (e ParseError) Error() string {
	return e.Msg
}

// List all legal cmds in a dockerfile
func AllCmds() []string {
	var ret []string
	for k := range command.Commands {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret
}

// Parse a Dockerfile from a reader.  A ParseError may occur.
func ParseReader(file io.Reader) ([]Command, error) {
	res, err := parser.Parse(file)
	if err != nil {
		return nil, ParseError{err.Error()}
	}

	var ret []Command
	for _, child := range res.AST.Children {
		cmd := Command{
			Cmd:       child.Value,
			Original:  child.Original,
			StartLine: child.StartLine,
			EndLine:   child.EndLine,
			Flags:     child.Flags,
		}

		// Only happens for ONBUILD
		if child.Next != nil && len(child.Next.Children) > 0 {
			cmd.SubCmd = child.Next.Children[0].Value
			child = child.Next.Children[0]
		}

		cmd.Json = child.Attributes["json"]
		for n := child.Next; n != nil; n = n.Next {
			cmd.Value = append(cmd.Value, n.Value)
		}

		if len(child.Heredocs) != 0 {
			// For heredocs, add heredocs extra lines to Original,
			// and to the heredocs list.
			cmd.Original = cmd.Original + "\n"
			for _, heredoc := range child.Heredocs {
				cmd.Original = cmd.Original + heredoc.Content + heredoc.Name + "\n"
				cmd.Heredocs = append(cmd.Heredocs, Heredoc{Name: heredoc.Name,
					FileDescriptor: heredoc.FileDescriptor,
					Content:        heredoc.Content})
			}
		}

		ret = append(ret, cmd)
	}
	return ret, nil
}

// Parse a Dockerfile from a filename.  An IOError or ParseError may occur.
func ParseFile(filename string) ([]Command, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, IOError{err.Error()}
	}
	defer file.Close()

	return ParseReader(file)
}
