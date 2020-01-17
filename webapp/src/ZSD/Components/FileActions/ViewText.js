"use strict;"

exports.highlightCode = function() {
    document.querySelectorAll("pre code").forEach((block) => {
	hljs.highlightBlock(block);
    });
}
