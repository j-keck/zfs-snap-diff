"use strict";

exports.setItem_ = function(k, v) {
    window.localStorage.setItem(k, v);
};

exports.getItem_ = function(k) {
    return window.localStorage.getItem(k);
}
