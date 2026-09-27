package dev

import (
	"fmt"
	"html"
)

func DefaultErrorPage(err error) []byte {
	return fmt.Appendf(nil,
		`<!DOCTYPE html>
		<html lang="en">
		<head>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<title>Build Error</title>
		<style>
		body { margin: 0; padding: 2rem; background: #1a0f0f; color: #ffb4b4;
				font: 14px/1.7 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
		h1 { color: #ff6b6b; font-size: 18px; margin: 0 0 1rem; }
		pre { background: #2a1414; padding: 1rem; border-radius: 6px;
				white-space: pre-wrap; word-break: break-word; margin: 0; }
		.hint { color: #888; margin-top: 1rem; font-size: 12px; }
		</style>
		</head>
		<body>
		<h1>Build Error</h1>
		<pre>%s</pre>
		<p class="hint">Fix the error and save. The page will refresh automatically.</p>
		</body>
		</html>`, html.EscapeString(err.Error()))
}
