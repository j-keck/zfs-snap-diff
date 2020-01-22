"use strict;"

exports.highlightCode = function() {
    let hljs = require('highlight.js');
    document.querySelectorAll("pre code").forEach((block) => {
	hljs.highlightBlock(block);
    });
}
