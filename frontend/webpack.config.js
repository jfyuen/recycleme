const path = require('path');

module.exports = {
    entry: {
        main: './app.js'
    },
    output: {
        path: path.resolve(__dirname, "../static/js"),
        publicPath: '../static/js',
        filename: 'bundle.js'
    },
    mode: 'production'
};