// SPDX-FileCopyrightText: 2025 Daniel Pfeifer <daniel@pfeifer-mail.de>
// SPDX-License-Identifier: ISC

package ctestxml

import (
	"bytes"
	"strings"
	"time"

	"github.com/purpleKarrot/cdash-proxy/algorithm"
	"github.com/purpleKarrot/cdash-proxy/ctestxml/buildparser"
	"github.com/purpleKarrot/cdash-proxy/model"
)

func parseBuild(build *Build) TimedCommands {
	ret := TimedCommands{
		StartTime: time.Unix(build.StartTime, 0),
		EndTime:   time.Unix(build.EndTime, 0),
	}

	ret.Commands = append(ret.Commands, model.Command{
		Role:         "Build",
		CommandLine:  build.Command,
		StartTime:    &ret.StartTime,
		Duration:     ret.EndTime.Sub(ret.StartTime).Milliseconds(),
		StdOut:       combineOutput(build.Diagnostics),
		Diagnostics:  mapDiagnostics(build.Diagnostics),
		Measurements: map[string]float64{},
	})

	for _, failure := range build.Failures {
		ret.Commands = append(ret.Commands, model.Command{
			CommandLine:      strings.Join(failure.Argv, " "),
			Result:           failure.ExitCondition,
			Role:             "Compile",
			Target:           failure.Target,
			TargetType:       failure.OutputType,
			TargetLabels:     failure.Labels,
			Source:           failure.SourceFile,
			Language:         failure.Language,
			StdOut:           buildparser.CleanOutput(failure.StdOut),
			StdErr:           buildparser.CleanOutput(failure.StdErr),
			Diagnostics:      buildparser.ParseOutput(failure.SourceFile, failure.StdErr),
			Attributes:       map[string]string{},
			WorkingDirectory: failure.WorkingDirectory,
			// Outputs:          failure.OutputFile,
		})
	}

	return ret
}

func combineOutput(messages []Diagnostic) string {
	var buffer bytes.Buffer
	for _, e := range messages {
		buffer.WriteString(e.PreContext)
		buffer.WriteString(e.Text)
		buffer.WriteString("\n")
		buffer.WriteString(e.PostContext)
	}
	return buffer.String()
}

func mapDiagnostics(messages []Diagnostic) []model.Diagnostic {
	return algorithm.Map(messages, func(e Diagnostic) model.Diagnostic {
		return model.Diagnostic{
			FilePath: e.SourceFile,
			Line:     e.SourceLine,
			Column:   -1,
			Type:     parseDiagnosticType(e.XMLName.Local),
			Message:  e.Text,
			Option:   "",
		}
	})
}

func parseDiagnosticType(s string) string {
	if s == "Error" {
		return "Error"
	}
	return "Warning"
}
