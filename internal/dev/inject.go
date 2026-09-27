package dev

import "bytes"

const devScriptTag = `<script src="/_dev/client.js" defer></script>`

// 在 </body> 前插入开发脚本
func InjectDevScript(html []byte) []byte {
	if bytes.Contains(html, []byte("/_dev/client.js")) { // 已注入则跳过
		return html
	}
	idx := bytes.LastIndex(html, []byte("</body>"))
	if idx < 0 {
		return append(html, []byte(devScriptTag)...)
	}
	out := make([]byte, 0, len(html)+len(devScriptTag))
	out = append(out, html[:idx]...)
	out = append(out, []byte(devScriptTag)...)
	out = append(out, html[idx:]...)
	return out
}

const DevClientJS = `
(function () {
	if (window.__ssg_dev__) return;
  	window.__ssg_dev__ = true;

	var seq = Date.now();
  	var es = new EventSource('/_dev/sse');

	es.onmessage = function (e) {
		var msg = e.data;
		if (msg === 'reload') {
			location.reload();
		} else if (msg === 'css') {
			var links = document.querySelectorAll('link[rel="stylesheet"]');
			for (var i = 0; i < links.length; i++) {
				var href = links[i].href.split('?')[0];
				links[i].href = href + '?t=' + (++seq);
			}
		}
	};

	es.onerror = function () {
		// EventSource 会自动重连，无需处理
	};
})();
`
