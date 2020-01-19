import 'bootstrap'
import 'bootstrap/dist/css/bootstrap.css'
import '@fortawesome/fontawesome-free/css/all.css'

// FIXME: highlighting does not work
import hljs from 'highlight.js/lib/index.js';
hljs.initHighlightingOnLoad();
import 'highlight.js/styles/default.css'

require("./output/Main/index.js").main();
