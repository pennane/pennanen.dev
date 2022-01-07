module.exports = {
    "parserOptions": {
        "ecmaVersion": 6,
        "sourceType": "module",
        "ecmaFeatures": {
            "jsx": true
        }
    },
    "rules": {
        // the rules below should be sorted in a same way they are sorted on http://eslint.org/docs/rules page
        // http://eslint.org/docs/rules/#possible-errors
        "no-caller": 2,
        "no-control-regex": 2,
        "no-empty": 0,
        "no-invalid-regexp": 2,
        "no-regex-spaces": 2,
        "no-unsafe-negation": 1,
        "valid-jsdoc": 0,
        "valid-typeof": 2,
        // http://eslint.org/docs/rules/#best-practices
        "curly": 2,
        "eqeqeq": [2, "smart"],
        "guard-for-in": 0,
        "no-else-return": 1,
        "no-fallthrough": 2,
        "no-invalid-this": 0,
        "no-iterator": 2,
        "no-loop-func": 2,
        "no-console": 0,
        "no-multi-str": 2,
        "no-new-func": 2,
        "no-new-wrappers": 2,
        "no-new": 2,
        "no-proto": 2,
        "no-redeclare": 1,
        "no-script-url": 2,
        "wrap-iife": [2, "outside"],
        // http://eslint.org/docs/rules/#strict-mode
        "strict": 2,
        // http://eslint.org/docs/rules/#variables
        "no-shadow-restricted-names": 2,
        "no-shadow": 1,
        "no-undef": 2,
        "no-unused-vars": [0, {"vars": "all", "args": "none"}],
        "no-use-before-define": 0,
        // http://eslint.org/docs/rules/#nodejs-and-commonjs
        "no-new-require": 2,
        // http://eslint.org/docs/rules/#stylistic-issues
        "block-spacing": 0,
        "brace-style": [1, "1tbs", { allowSingleLine: true }],
        "camelcase": 1,
        "comma-dangle": 2,
        "comma-spacing": 0,
        "comma-style": [1, "last"],
        "computed-property-spacing": 1,
        "eol-last": 0,
        "func-call-spacing": 1,
       // "indent": [1, 4],
        "key-spacing": [1, { beforeColon: false, afterColon: true }],
        "new-cap": [0, {
            "capIsNewExceptions": [
                "$.Deferred",
                "$.Event",
                "CodeMirror.Pos",
                "Immutable.Map",
                "Immutable.List",
                "JSLINT"
            ]
        }],
        "new-parens": 2,
        "no-bitwise": 2,
        "no-new-object": 2,
        "no-trailing-spaces": 0,
        "semi-spacing": 0,
        "semi": 0
    },
    "env": {
        "browser": true,
        "node": true,
        "jasmine": true
    },
    "globals": {
        "audio": true,
        "Cookies": false,
        "$": false,
        "brackets": false,
        "clearTimeout": false,
        "console": false,
        "define": false,
        "require": false,
        "setTimeout": false,
        "window": false,
        "ArrayBuffer": false,
        "Uint32Array": false,
        "WebSocket": false,
        "XMLHttpRequest": false
    }
};