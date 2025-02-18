// SPDX-FileCopyrightText: 2025 Daniel Pfeifer <daniel@pfeifer-mail.de>
// SPDX-License-Identifier: ISC

package configure

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/purpleKarrot/cdash-proxy/model"
)

var cfgDiagRegex = regexp.MustCompile(`CMake (Deprecation Warning|Error|Warning \(dev\)|Warning) at ([^:]+):([0-9]+) \((.*)\):`)

func Parse(log string, result int) []model.Diagnostic {
	var diags []model.Diagnostic
	diag := model.Diagnostic{} // TODO: set to nil

	for _, line := range strings.Split(log, "\n") {
		if len(line) == 0 {
			diag.Message += "\n"
			continue
		} else if strings.HasPrefix(line, "  ") {
			diag.Message += line[2:] + "\n"
			continue
		} else if len(diag.Message) != 0 && len(diag.FilePath) != 0 {
			diag.Message = strings.TrimRight(diag.Message, "\n")
			diags = append(diags, diag)
			diag = model.Diagnostic{} // TODO: set to nil
		}

		if match := cfgDiagRegex.FindStringSubmatch(line); match != nil {
			linenr, _ := strconv.Atoi(match[3])
			diag = model.Diagnostic{
				FilePath: match[2],
				Line:     linenr,
				Column:   -1,
				Type:     cmakeDiagnosticType(match[1]),
				Option:   match[4],
			}
		}
	}

	if len(diags) == 0 && result != 0 {
		diags = append(diags, model.Diagnostic{
			FilePath: "CMakeLists.txt",
			Line:     -1,
			Column:   -1,
			Type:     "Error",
			Message:  fmt.Sprintf("Command finished with exit code %d", result),
		})
	}

	return diags
}

func cmakeDiagnosticType(s string) string {
	if s == "Error" {
		return "Error"
	}
	return "Warning"
}
